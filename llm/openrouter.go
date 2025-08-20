package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
	"unicode"
	
	"github.com/cohesion-org/deepseek-go/constants"
	openrouter "github.com/revrost/go-openrouter"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
)

type AIRouterReq struct {
	ToolCall           []openrouter.ToolCall
	ToolMessage        []openrouter.ChatCompletionMessage
	CurrentToolMessage []openrouter.ChatCompletionMessage
	
	OpenRouterMsgs []openrouter.ChatCompletionMessage
}

func (d *AIRouterReq) GetModel(l *LLM) {
	l.Model = param.DeepseekDeepseekR1_0528Free
	userInfo, err := db.GetUserByID(l.UserId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.OpenRouterModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (d *AIRouterReq) GetMessages(userId string, prompt string) {
	messages := make([]openrouter.ChatCompletionMessage, 0)
	
	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		aqs := msgRecords.AQs
		if len(aqs) > db.MaxQAPair {
			aqs = aqs[len(aqs)-db.MaxQAPair:]
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
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	d.GetModel(l)
	
	// set deepseek proxy
	config := openrouter.DefaultConfig(*conf.BaseConfInfo.OpenRouterToken)
	config.HTTPClient = utils.GetLLMProxyClient()
	client := openrouter.NewClientWithConfig(*config)
	
	request := openrouter.ChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: &openrouter.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.LLMConfInfo.MaxTokens,
		TopP:             float32(*conf.LLMConfInfo.TopP),
		FrequencyPenalty: float32(*conf.LLMConfInfo.FrequencyPenalty),
		TopLogProbs:      *conf.LLMConfInfo.TopLogProbs,
		LogProbs:         *conf.LLMConfInfo.LogProbs,
		Stop:             conf.LLMConfInfo.Stop,
		PresencePenalty:  float32(*conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(*conf.LLMConfInfo.Temperature),
		Tools:            l.OpenRouterTools,
	}
	
	request.Messages = d.OpenRouterMsgs
	
	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
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
			logger.Info("Stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.Warn("Stream error", "updateMsgID", l.MsgId, "err", err)
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
						logger.Error("requestToolsCall error", "updateMsgID", l.MsgId, "err", err)
					}
				}
			}
			
			if len(choice.Delta.Content) > 0 {
				msgInfoContent = l.SendMsg(msgInfoContent, choice.Delta.Content)
			}
		}
		
		if response.Usage != nil {
			l.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(l.Token))
		}
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 {
		l.MessageChan <- msgInfoContent
	}
	
	if !hasTools || len(d.CurrentToolMessage) == 0 {
		db.InsertMsgRecord(l.UserId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Token:    l.Token,
			Mode:     l.Model,
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

func (d *AIRouterReq) GetUserMessage(msg string) {
	d.GetMessage(openrouter.ChatMessageRoleUser, msg)
}

func (d *AIRouterReq) GetAssistantMessage(msg string) {
	d.GetMessage(openrouter.ChatMessageRoleAssistant, msg)
}

func (d *AIRouterReq) AppendMessages(client LLMClient) {
	if len(d.OpenRouterMsgs) == 0 {
		d.OpenRouterMsgs = make([]openrouter.ChatCompletionMessage, 0)
	}
	
	d.OpenRouterMsgs = append(d.OpenRouterMsgs, client.(*AIRouterReq).OpenRouterMsgs...)
}

func (d *AIRouterReq) GetMessage(role, msg string) {
	if len(d.OpenRouterMsgs) == 0 {
		d.OpenRouterMsgs = []openrouter.ChatCompletionMessage{
			{
				Role: role,
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
		return
	}
	
	d.OpenRouterMsgs = append(d.OpenRouterMsgs, openrouter.ChatCompletionMessage{
		Role: role,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type: openrouter.ChatMessagePartTypeText,
					Text: msg,
				},
			},
		},
	})
}

func (d *AIRouterReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	d.GetModel(l)
	config := openrouter.DefaultConfig(*conf.BaseConfInfo.OpenRouterToken)
	config.HTTPClient = utils.GetLLMProxyClient()
	client := openrouter.NewClientWithConfig(*config)
	
	request := openrouter.ChatCompletionRequest{
		Model:            l.Model,
		MaxTokens:        *conf.LLMConfInfo.MaxTokens,
		TopP:             float32(*conf.LLMConfInfo.TopP),
		FrequencyPenalty: float32(*conf.LLMConfInfo.FrequencyPenalty),
		TopLogProbs:      *conf.LLMConfInfo.TopLogProbs,
		LogProbs:         *conf.LLMConfInfo.LogProbs,
		Stop:             conf.LLMConfInfo.Stop,
		PresencePenalty:  float32(*conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(*conf.LLMConfInfo.Temperature),
		Tools:            l.OpenRouterTools,
		Messages:         d.OpenRouterMsgs,
	}
	
	// assign task
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("CreateChatCompletion error", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	l.Token += response.Usage.TotalTokens
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.OpenRouterMsgs[len(d.OpenRouterMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls)
	}
	
	return response.Choices[0].Message.Content.Text, nil
}

func (d *AIRouterReq) requestOneToolsCall(ctx context.Context, toolsCall []openrouter.ToolCall) {
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			return
		}
		
		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return
		}
		
		d.OpenRouterMsgs = append(d.OpenRouterMsgs, openrouter.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: openrouter.Content{
				Text: toolsData,
			},
			ToolCallID: tool.ID,
		})
		logger.Info("exec tool", "name", tool.Function.Name, "toolsData", toolsData)
	}
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
		
		if toolCall.Function.Arguments != "" && toolCall.Function.Name == "" {
			d.ToolCall[len(d.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}
		
		err := json.Unmarshal([]byte(d.ToolCall[len(d.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}
		
		mc, err := clients.GetMCPClientByToolName(d.ToolCall[len(d.ToolCall)-1].Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err, "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
				"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments)
			return err
		}
		
		toolsData, err := mc.ExecTools(ctx, d.ToolCall[len(d.ToolCall)-1].Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err, "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
				"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments)
			return err
		}
		d.CurrentToolMessage = append(d.CurrentToolMessage, openrouter.ChatCompletionMessage{
			Role: constants.ChatMessageRoleTool,
			Content: openrouter.Content{
				Text: toolsData,
			},
			ToolCallID: d.ToolCall[len(d.ToolCall)-1].ID,
		})
		
		logger.Info("send tool request", "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
			"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments,
			"res", toolsData)
	}
	
	return nil
}
