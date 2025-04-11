package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

const (
	OneMsgLen       = 3896
	FirstSendLen    = 30
	NonFirstSendLen = 500
)

var (
	toolsJsonErr = errors.New("tools json error")
)

type Deepseek interface {
	GetContent()
}

type DeepseekReq struct {
	MessageChan chan *param.MsgInfo
	Update      tgbotapi.Update
	Bot         *tgbotapi.BotAPI
	Content     string
	Model       string
	ToolCall    deepseek.ToolCall
}

// GetContent get comment from deepseek
func (d *DeepseekReq) GetContent() {
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

	if d.Content == "" {
		logger.Warn("content empty")
		return
	}

	text := strings.ReplaceAll(d.Content, "@"+d.Bot.Self.UserName, "")
	err := d.callDeepSeekAPI(text)
	if err != nil {
		logger.Error("Error calling DeepSeek API", "err", err)
	}
}

// callDeepSeekAPI request DeepSeek API and get response
func (d *DeepseekReq) callDeepSeekAPI(prompt string) error {
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
				logger.Info("context content", "dialog", i, "question:", record.Question, "answer:", record.Answer)
				messages = append(messages, deepseek.ChatCompletionMessage{
					Role:    constants.ChatMessageRoleAssistant,
					Content: record.Answer,
				})
				messages = append(messages, deepseek.ChatCompletionMessage{
					Role:    constants.ChatMessageRoleUser,
					Content: record.Question,
				})
			}
		}
	}

	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: prompt,
	})

	logger.Info("msg receive", "userID", userId, "prompt", prompt)

	return d.send(messages)
}

func (d *DeepseekReq) send(messages []deepseek.ChatCompletionMessage) error {
	start := time.Now()
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(d.Update)
	// set deepseek proxy
	httpClient := &http.Client{
		Timeout: 30 * time.Minute,
	}

	if *conf.DeepseekProxy != "" {
		proxy, err := url.Parse(*conf.DeepseekProxy)
		if err != nil {
			logger.Error("parse deepseek proxy error", "err", err)
		} else {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	client, err := deepseek.NewClientWithOptions(*conf.DeepseekToken,
		deepseek.WithBaseURL(*conf.CustomUrl), deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.Error("Error creating deepseek client", "err", err)
		return err
	}

	request := &deepseek.StreamChatCompletionRequest{
		Model:  d.Model,
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

	ctx := context.Background()

	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", updateMsgID, "err", err)
		return err
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}

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
				err = d.requestToolsCall(ctx, choice, messages)
				if err != nil {
					if errors.Is(err, toolsJsonErr) {
						continue
					} else {
						logger.Warn("requestToolsCall error", "updateMsgID", updateMsgID, "err", err)
						return err
					}
				}
				return nil
			}

			// exceed max telegram one message length
			if utils.Utf16len(msgInfoContent.Content) > OneMsgLen {
				d.MessageChan <- msgInfoContent
				msgInfoContent = &param.MsgInfo{
					SendLen:     NonFirstSendLen,
					FullContent: msgInfoContent.FullContent,
					Token:       msgInfoContent.Token,
				}
			}

			msgInfoContent.Content += choice.Delta.Content
			msgInfoContent.FullContent += choice.Delta.Content
			if len(msgInfoContent.Content) > msgInfoContent.SendLen {
				d.MessageChan <- msgInfoContent
				msgInfoContent.SendLen += NonFirstSendLen
			}
		}

		if response.Usage != nil {
			msgInfoContent.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(msgInfoContent.Token))
		}
	}

	d.MessageChan <- msgInfoContent

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (d *DeepseekReq) requestToolsCall(ctx context.Context, choice deepseek.StreamChoices, messages []deepseek.ChatCompletionMessage) error {

	toolMessage := make([]deepseek.ChatCompletionMessage, 0)
	for _, toolCall := range choice.Delta.ToolCalls {
		property := make(map[string]interface{})

		if toolCall.Function.Name != "" {
			d.ToolCall.Function.Name = toolCall.Function.Name
		}

		if toolCall.ID != "" {
			d.ToolCall.ID = toolCall.ID
		}

		if toolCall.Type != "" {
			d.ToolCall.Type = toolCall.Type
		}

		if toolCall.Function.Arguments != "" {
			d.ToolCall.Function.Arguments += toolCall.Function.Arguments
		}

		err := json.Unmarshal([]byte(d.ToolCall.Function.Arguments), &property)
		if err != nil {
			return toolsJsonErr
		}

		mc, err := clients.GetMCPClientByToolName(d.ToolCall.Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return err
		}

		toolsData, err := mc.ExecTools(ctx, d.ToolCall.Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return err
		}
		toolMessage = append(toolMessage, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: d.ToolCall.ID,
		})

	}

	logger.Info("send tool request", "function", d.ToolCall.Function.Name, "toolCall", d.ToolCall.ID, "argument", d.ToolCall.Function.Arguments)
	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:      deepseek.ChatMessageRoleAssistant,
		Content:   "",
		ToolCalls: []deepseek.ToolCall{d.ToolCall},
	})
	messages = append(messages, toolMessage...)

	return d.send(messages)

}

// GetBalanceInfo get balance info
func GetBalanceInfo() *deepseek.BalanceResponse {
	client := deepseek.NewClient(*conf.DeepseekToken)
	ctx := context.Background()
	balance, err := deepseek.GetBalance(client, ctx)
	if err != nil {
		logger.Error("Error getting balance", "err", err)
	}

	if balance == nil || len(balance.BalanceInfos) == 0 {
		logger.Error("No balance information returned")
	}

	return balance
}
