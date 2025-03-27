package deepseek

import (
	"context"
	"errors"
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

// GetContentFromDP get comment from deepseek
func GetContentFromDP(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	text := strings.ReplaceAll(content, "@"+bot.Self.UserName, "")
	err := callDeepSeekAPI(text, update, messageChan)
	if err != nil {
		logger.Error("Error calling DeepSeek API", "err", err)
	}
	close(messageChan)
}

// callDeepSeekAPI request DeepSeek API and get response
func callDeepSeekAPI(prompt string, update tgbotapi.Update, messageChan chan *param.MsgInfo) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	model := deepseek.DeepSeekChat
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		model = userInfo.Mode
	}

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
		Model:  model,
		Stream: true,
		StreamOptions: deepseek.StreamOptions{
			IncludeUsage: true,
		},
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

	request.Messages = messages

	ctx := context.Background()

	logger.Info("msg receive", "userID", userId, "prompt", prompt)
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
			// exceed max telegram one message length
			if utils.Utf16len(msgInfoContent.Content) > OneMsgLen {
				messageChan <- msgInfoContent
				msgInfoContent = &param.MsgInfo{
					SendLen:     NonFirstSendLen,
					FullContent: msgInfoContent.FullContent,
					Token:       msgInfoContent.Token,
				}
			}

			msgInfoContent.Content += choice.Delta.Content
			msgInfoContent.FullContent += choice.Delta.Content
			if len(msgInfoContent.Content) > msgInfoContent.SendLen {
				messageChan <- msgInfoContent
				msgInfoContent.SendLen += NonFirstSendLen
			}
		}

		if response.Usage != nil {
			msgInfoContent.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(msgInfoContent.Token))
		}
	}

	messageChan <- msgInfoContent

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
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
