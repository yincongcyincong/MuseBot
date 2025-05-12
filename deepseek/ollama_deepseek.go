package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type OllamaDeepseekReq struct {
	MessageChan chan *param.MsgInfo
	Update      tgbotapi.Update
	Bot         *tgbotapi.BotAPI
	Content     string
	Model       string
	Token       int

	ToolCall           []deepseek.ToolCall
	DeepSeekContent    string
	ToolMessage        []deepseek.ChatCompletionMessage
	CurrentToolMessage []deepseek.ChatCompletionMessage
}

// GetContent get comment from deepseek
func (d *OllamaDeepseekReq) GetContent() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// check user chat exceed max count
	if utils.CheckUserChatExceed(d.Update, d.Bot) {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Error("GetContent panic err", "err", err)
		}
		utils.DecreaseUserChat(d.Update)
		close(d.MessageChan)
	}()

	if d.Content == "" && d.Update.Message.Voice != nil && *conf.AudioAppID != "" {
		audioContent := utils.GetAudioContent(d.Update, d.Bot)
		if audioContent == nil {
			logger.Warn("audio url empty")
			return
		}
		d.Content = FileRecognize(audioContent)
	}

	if d.Content == "" && d.Update.Message.Photo != nil {
		imageContent, err := GetImageContent(utils.GetPhotoContent(d.Update, d.Bot))
		if err != nil {
			logger.Warn("get image content err", "err", err)
			return
		}
		d.Content = imageContent
	}

	if d.Content == "" {
		logger.Warn("content empty")
		return
	}

	text := strings.ReplaceAll(d.Content, "@"+d.Bot.Self.UserName, "")
	err := d.callDeepSeekAPI(ctx, text)
	if err != nil {
		logger.Error("Error calling DeepSeek API", "err", err)
	}
}

// callDeepSeekAPI request DeepSeek API and get response
func (d *OllamaDeepseekReq) callDeepSeekAPI(ctx context.Context, prompt string) error {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(d.Update)
	d.Model = deepseek.DeepSeekChat
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		d.Model = userInfo.Mode
	}

	messages := make([]deepseek.ChatCompletionMessage, 0)

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
				messages = append(messages, deepseek.ChatCompletionMessage{
					Role:    constants.ChatMessageRoleUser,
					Content: record.Question,
				})
				if record.Content != "" {
					toolsMsgs := make([]deepseek.ChatCompletionMessage, 0)
					err = json.Unmarshal([]byte(record.Content), &toolsMsgs)
					if err != nil {
						logger.Error("Error unmarshalling tools json", "err", err)
					} else {
						messages = append(messages, toolsMsgs...)
					}
				}
				messages = append(messages, deepseek.ChatCompletionMessage{
					Role:    constants.ChatMessageRoleAssistant,
					Content: record.Answer,
				})
			}
		}
	}

	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: prompt,
	})

	logger.Info("msg receive", "userID", userId, "prompt", prompt)

	return d.send(ctx, messages)
}

func (d *OllamaDeepseekReq) send(ctx context.Context, messages []deepseek.ChatCompletionMessage) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(d.Update)

	request := &deepseek.StreamChatCompletionRequest{
		Model:  "llava:latest",
		Stream: true,
		StreamOptions: deepseek.StreamOptions{
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
		Tools:            conf.DeepseekTools,
	}

	request.Messages = messages

	stream, err := deepseek.CreateOllamaChatCompletionStream(ctx, request)
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
					if errors.Is(err, toolsJsonErr) {
						continue
					} else {
						logger.Error("requestToolsCall error", "updateMsgID", updateMsgID, "err", err)
					}
				}
			}

			if !hasTools {
				d.sendMsg(msgInfoContent, choice)
			}
		}

		if response.Usage != nil {
			d.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(d.Token))
		}
	}

	if !hasTools || len(d.CurrentToolMessage) == 0 {
		d.MessageChan <- msgInfoContent

		data, _ := json.Marshal(d.ToolMessage)
		db.InsertMsgRecord(userId, &db.AQ{
			Question: d.Content,
			Answer:   d.DeepSeekContent,
			Content:  string(data),
			Token:    d.Token,
		}, true)
	} else {
		d.CurrentToolMessage = append([]deepseek.ChatCompletionMessage{
			{
				Role:      deepseek.ChatMessageRoleAssistant,
				Content:   d.DeepSeekContent,
				ToolCalls: d.ToolCall,
			},
		}, d.CurrentToolMessage...)

		d.ToolMessage = append(d.ToolMessage, d.CurrentToolMessage...)
		messages = append(messages, d.CurrentToolMessage...)
		d.CurrentToolMessage = make([]deepseek.ChatCompletionMessage, 0)
		d.ToolCall = make([]deepseek.ToolCall, 0)
		return d.send(ctx, messages)
	}

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (d *OllamaDeepseekReq) sendMsg(msgInfoContent *param.MsgInfo, choice deepseek.StreamChoices) {
	// exceed max telegram one message length
	if utils.Utf16len(msgInfoContent.Content) > OneMsgLen {
		d.MessageChan <- msgInfoContent
		msgInfoContent = &param.MsgInfo{
			SendLen: NonFirstSendLen,
		}
	}

	msgInfoContent.Content += choice.Delta.Content
	d.DeepSeekContent += choice.Delta.Content
	if len(msgInfoContent.Content) > msgInfoContent.SendLen {
		d.MessageChan <- msgInfoContent
		msgInfoContent.SendLen += NonFirstSendLen
	}
}

func (d *OllamaDeepseekReq) requestToolsCall(ctx context.Context, choice deepseek.StreamChoices) error {

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
			return toolsJsonErr
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
		d.CurrentToolMessage = append(d.CurrentToolMessage, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: d.ToolCall[len(d.ToolCall)-1].ID,
		})

	}

	logger.Info("send tool request", "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
		"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments)

	return nil

}
