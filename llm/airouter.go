package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/cohesion-org/deepseek-go/constants"
	openrouter "github.com/revrost/go-openrouter"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type AIRouterReq struct {
	ToolCall           []openrouter.ToolCall
	ToolMessage        []openrouter.ChatCompletionMessage
	CurrentToolMessage []openrouter.ChatCompletionMessage

	OpenRouterMsgs []openrouter.ChatCompletionMessage
}

// CallLLMAPI request DeepSeek API and get response
func (d *AIRouterReq) CallLLMAPI(ctx context.Context, prompt string, l *LLM) error {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)

	d.GetMessages(userId, prompt)

	logger.Info("msg receive", "userID", userId, "prompt", prompt)

	return d.Send(ctx, l)
}

func (d *AIRouterReq) GetModel(l *LLM) {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	l.Model = param.DeepseekDeepseekR1_0528Free
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.OpenRouterModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (d *AIRouterReq) GetMessages(userId int64, prompt string) {
	messages := make([]openrouter.ChatCompletionMessage, 0)

	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		aqs := msgRecords.AQs
		if len(aqs) > 10 {
			aqs = aqs[len(aqs)-10:]
		}

		for i, record := range aqs {
			if record.Answer != "" && record.Question != "" {
				logger.Info("context content", "dialog", i, "question:", record.Question,
					"toolContent", record.Content, "answer:", record.Answer)
				messages = append(messages, openrouter.ChatCompletionMessage{
					Role: constants.ChatMessageRoleUser,
					Content: openrouter.Content{
						Multi: []openrouter.ChatMessagePart{
							{
								Type: openrouter.ChatMessagePartTypeText,
								Text: record.Question,
							},
						},
					},
				})
				if record.Content != "" {
					toolsMsgs := make([]openrouter.ChatCompletionMessage, 0)
					err := json.Unmarshal([]byte(record.Content), &toolsMsgs)
					if err != nil {
						logger.Error("Error unmarshalling tools json", "err", err)
					} else {
						messages = append(messages, toolsMsgs...)
					}
				}
				messages = append(messages, openrouter.ChatCompletionMessage{
					Role: constants.ChatMessageRoleAssistant,
					Content: openrouter.Content{
						Multi: []openrouter.ChatMessagePart{
							{
								Type: openrouter.ChatMessagePartTypeText,
								Text: record.Answer,
							},
						},
					},
				})
			}
		}
	}
	messages = append(messages, openrouter.ChatCompletionMessage{
		Role: constants.ChatMessageRoleUser,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type: openrouter.ChatMessagePartTypeText,
					Text: prompt,
				},
			},
		},
	})

	d.OpenRouterMsgs = messages
}

func (d *AIRouterReq) Send(ctx context.Context, l *LLM) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	d.GetModel(l)

	// set deepseek proxy
	config := openrouter.DefaultConfig(*conf.OpenRouterToken)
	config.HTTPClient = utils.GetDeepseekProxyClient()
	client := openrouter.NewClientWithConfig(*config)

	request := openrouter.ChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: &openrouter.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
		Tools:            l.OpenRouterTools,
	}

	request.Messages = d.OpenRouterMsgs

	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", updateMsgID, "err", err)
		return err
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}

	hasTools := false
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.Info("Stream finished", "updateMsgID", updateMsgID)
			break
		}
		if err != nil {
			logger.Warn("Stream error", "updateMsgID", updateMsgID, "err", err)
			break
		}
		for _, choice := range response.Choices {
			if len(choice.Delta.ToolCalls) > 0 {
				hasTools = true
				err = d.requestToolsCall(ctx, choice)
				if err != nil {
					if errors.Is(err, ToolsJsonErr) {
						continue
					} else {
						logger.Error("requestToolsCall error", "updateMsgID", updateMsgID, "err", err)
					}
				}
			}

			if !hasTools {
				msgInfoContent = l.sendMsg(msgInfoContent, choice.Delta.Content)
			}
		}

		if response.Usage != nil {
			l.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(l.Token))
		}
	}

	if !hasTools || len(d.CurrentToolMessage) == 0 {
		l.MessageChan <- msgInfoContent

		data, _ := json.Marshal(d.ToolMessage)
		db.InsertMsgRecord(userId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Content:  string(data),
			Token:    l.Token,
		}, true)
	} else {
		d.CurrentToolMessage = append([]openrouter.ChatCompletionMessage{
			{
				Role: openrouter.ChatMessageRoleAssistant,
				Content: openrouter.Content{
					Text: l.WholeContent,
				},
				ToolCalls: d.ToolCall,
			},
		}, d.CurrentToolMessage...)

		d.ToolMessage = append(d.ToolMessage, d.CurrentToolMessage...)
		d.OpenRouterMsgs = append(d.OpenRouterMsgs, d.CurrentToolMessage...)
		d.CurrentToolMessage = make([]openrouter.ChatCompletionMessage, 0)
		d.ToolCall = make([]openrouter.ToolCall, 0)
		return d.Send(ctx, l)
	}

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (d *AIRouterReq) GetMessage(msg string) {
	d.OpenRouterMsgs = []openrouter.ChatCompletionMessage{
		{
			Role: constants.ChatMessageRoleAssistant,
			Content: openrouter.Content{
				Multi: []openrouter.ChatMessagePart{
					{
						Type: openrouter.ChatMessagePartTypeText,
						Text: msg,
					},
				},
			},
		},
	}
}

func (d *AIRouterReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(l.Update)

	d.GetModel(l)
	config := openrouter.DefaultConfig(*conf.OpenRouterToken)
	config.HTTPClient = utils.GetDeepseekProxyClient()
	client := openrouter.NewClientWithConfig(*config)

	request := openrouter.ChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: &openrouter.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
		Tools:            l.OpenRouterTools,
	}

	// assign task
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", updateMsgID, "err", err)
		return "", err
	}

	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}

	l.Token += response.Usage.TotalTokens

	return response.Choices[0].Message.Content.Text, nil
}

func (d *AIRouterReq) requestToolsCall(ctx context.Context, choice openrouter.ChatCompletionStreamChoice) error {

	for _, toolCall := range choice.Delta.ToolCalls {
		property := make(map[string]interface{})

		if toolCall.Function.Name != "" {
			d.ToolCall = append(d.ToolCall, toolCall)
			d.ToolCall[len(d.ToolCall)-1].Function.Name = toolCall.Function.Name
		}

		if toolCall.ID != "" {
			d.ToolCall[len(d.ToolCall)-1].ID = toolCall.ID
		}

		if toolCall.Type != "" {
			d.ToolCall[len(d.ToolCall)-1].Type = toolCall.Type
		}

		if toolCall.Function.Arguments != "" {
			d.ToolCall[len(d.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}

		err := json.Unmarshal([]byte(d.ToolCall[len(d.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}

		mc, err := clients.GetMCPClientByToolName(d.ToolCall[len(d.ToolCall)-1].Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return err
		}

		toolsData, err := mc.ExecTools(ctx, d.ToolCall[len(d.ToolCall)-1].Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return err
		}
		d.CurrentToolMessage = append(d.CurrentToolMessage, openrouter.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: openrouter.Content{
				Text: toolsData,
			},
			ToolCallID: d.ToolCall[len(d.ToolCall)-1].ID,
		})

	}

	logger.Info("send tool request", "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
		"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments)

	return nil
}
