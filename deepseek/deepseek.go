package deepseek

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

const (
	OneMsgLen       = 3896
	FirstSendLen    = 30
	NonFirstSendLen = 300
)

// GetContentFromDP get comment from deepseek
func GetContentFromDP(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	text := strings.ReplaceAll(content, "@"+bot.Self.UserName, "")
	err := callDeepSeekAPI(text, update, messageChan)
	if err != nil {
		log.Printf("Error calling DeepSeek API: %s\n", err)
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
		log.Printf("Error getting user info: %s\n", err)
	}
	if userInfo != nil && userInfo.Mode != "" {
		log.Printf("User info: %d, %s\n", userInfo.UserId, userInfo.Mode)
		model = userInfo.Mode
	}

	client := deepseek.NewClient(*conf.DeepseekToken, *conf.CustomUrl)
	request := &deepseek.StreamChatCompletionRequest{
		Model:  model,
		Stream: true,
	}
	messages := make([]deepseek.ChatCompletionMessage, 0)

	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		for _, record := range msgRecords.AQs {
			if record.Answer != "" && record.Question != "" {
				log.Println("question:", record.Question, "answer:", record.Answer)
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

	fmt.Printf("[%d]: %s\n", userId, prompt)
	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		log.Printf("ChatCompletionStream error: %d, %v\n", updateMsgID, err)
		return err
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Printf("\n %d Stream finished", updateMsgID)
			break
		}
		if err != nil {
			fmt.Printf("\n %d Stream error: %v\n", updateMsgID, err)
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
		log.Printf("Error getting balance: %v\n", err)
	}

	if balance == nil || len(balance.BalanceInfos) == 0 {
		log.Printf("No balance information returned\n")
	}

	return balance
}
