package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
	"google.golang.org/genai"
)

type GeminiReq struct {
	ToolCall           []*genai.FunctionCall
	ToolMessage        []*genai.Content
	CurrentToolMessage []*genai.Content
}

func (h *GeminiReq) CallLLMAPI(ctx context.Context, prompt string, l *LLM) error {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)

	l.Model = param.ModelGemini20Flash
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.GeminiModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}

	messages := h.getMessages(userId)

	logger.Info("msg receive", "userID", userId, "prompt", l.Content)
	return h.send(ctx, messages, prompt, l)
}

func (h *GeminiReq) getMessages(userId int64) []*genai.Content {
	messages := make([]*genai.Content, 0)

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

				messages = append(messages, &genai.Content{
					Role: genai.RoleUser,
					Parts: []*genai.Part{
						{
							Text: record.Question,
						},
					},
				})

				if record.Content != "" {
					toolsMsgs := make([]*genai.Content, 0)
					err := json.Unmarshal([]byte(record.Content), &toolsMsgs)
					if err != nil {
						logger.Error("Error unmarshalling tools json", "err", err)
					} else {
						messages = append(messages, toolsMsgs...)
					}
				}

				messages = append(messages, &genai.Content{
					Role: genai.RoleModel,
					Parts: []*genai.Part{
						{
							Text: record.Answer,
						},
					},
				})

			}
		}
	}

	return messages
}

func (h *GeminiReq) send(ctx context.Context, messages []*genai.Content, prompt string, l *LLM) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	httpClient := utils.GetDeepseekProxyClient()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     *conf.GeminiToken,
	})
	if err != nil {
		logger.Error("init gemini client fail", "err", err)
		return err
	}

	config := &genai.GenerateContentConfig{
		TopP:             genai.Ptr[float32](float32(*conf.TopP)),
		FrequencyPenalty: genai.Ptr[float32](float32(*conf.FrequencyPenalty)),
		PresencePenalty:  genai.Ptr[float32](float32(*conf.PresencePenalty)),
		Temperature:      genai.Ptr[float32](float32(*conf.Temperature)),
		Tools:            conf.GeminiTools,
	}

	chat, err := client.Chats.Create(ctx, l.Model, config, messages)
	if err != nil {
		logger.Error("create chat fail", "err", err)
		return err
	}

	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}

	hasTools := false
	for response, err := range chat.SendMessageStream(ctx, *genai.NewPartFromText(prompt)) {
		if errors.Is(err, io.EOF) {
			logger.Info("stream finished", "updateMsgID", updateMsgID)
			break
		}
		if err != nil {
			logger.Error("stream error:", "updateMsgID", updateMsgID, "err", err)
			break
		}

		response.Text()
		toolCalls := response.FunctionCalls()
		if len(toolCalls) > 0 {
			hasTools = true
			err = h.requestToolsCall(ctx, response)
			if err != nil {
				if errors.Is(err, ToolsJsonErr) {
					continue
				} else {
					logger.Error("requestToolsCall error", "updateMsgID", updateMsgID, "err", err)
				}
			}
		}

		if !hasTools {
			msgInfoContent = l.sendMsg(msgInfoContent, response.Text())
		}

		if response.UsageMetadata != nil {
			l.Token += int(response.UsageMetadata.TotalTokenCount)
			metrics.TotalTokens.Add(float64(l.Token))
		}

	}

	if !hasTools || len(h.CurrentToolMessage) == 0 {
		l.MessageChan <- msgInfoContent

		data, _ := json.Marshal(h.ToolMessage)
		db.InsertMsgRecord(userId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Content:  string(data),
			Token:    l.Token,
		}, true)
	} else {
		h.ToolMessage = append(h.ToolMessage, h.CurrentToolMessage...)
		messages = append(messages, h.CurrentToolMessage...)
		h.CurrentToolMessage = make([]*genai.Content, 0)
		h.ToolCall = make([]*genai.FunctionCall, 0)
		return h.send(ctx, messages, prompt, l)
	}

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (h *GeminiReq) requestToolsCall(ctx context.Context, response *genai.GenerateContentResponse) error {

	for _, toolCall := range response.FunctionCalls() {

		if toolCall.Name != "" {
			h.ToolCall = append(h.ToolCall, toolCall)
			h.ToolCall[len(h.ToolCall)-1].Name = toolCall.Name
		}

		if toolCall.ID != "" {
			h.ToolCall[len(h.ToolCall)-1].ID = toolCall.ID
		}

		if toolCall.Args != nil {
			h.ToolCall[len(h.ToolCall)-1].Args = toolCall.Args
		}

		mc, err := clients.GetMCPClientByToolName(h.ToolCall[len(h.ToolCall)-1].Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return err
		}

		toolsData, err := mc.ExecTools(ctx, h.ToolCall[len(h.ToolCall)-1].Name, h.ToolCall[len(h.ToolCall)-1].Args)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return err
		}
		h.CurrentToolMessage = append(h.CurrentToolMessage, &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{
				{
					FunctionCall: toolCall,
				},
			},
		})

		h.CurrentToolMessage = append(h.CurrentToolMessage, &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{
				{
					FunctionResponse: &genai.FunctionResponse{
						Response: map[string]any{"output": toolsData},
						ID:       h.ToolCall[len(h.ToolCall)-1].ID,
						Name:     h.ToolCall[len(h.ToolCall)-1].Name,
					},
				},
			},
		})

	}

	logger.Info("send tool request", "function", h.ToolCall[len(h.ToolCall)-1].Name,
		"toolCall", h.ToolCall[len(h.ToolCall)-1].ID, "argument", h.ToolCall[len(h.ToolCall)-1].Args)

	return nil
}
