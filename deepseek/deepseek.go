package deepseek

import (
	"context"
	"errors"
	"fmt"
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"io"
	"log"
	"strings"
)

const (
	OneMsgLen       = 1500
	FirstSendLen    = 30
	NonFirstSendLen = 300
)

func GetContentFromDP(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	text := strings.ReplaceAll(update.Message.Text, "@"+bot.Self.UserName, "")
	err := callDeepSeekAPI(text, update, messageChan)
	if err != nil {
		log.Printf("Error calling DeepSeek API: %s\n", err)
	}
	close(messageChan)
}

// callDeepSeekAPI request DeepSeek API and get response
func callDeepSeekAPI(prompt string, update tgbotapi.Update, messageChan chan *param.MsgInfo) error {
	updateMsgID := update.Message.MessageID
	model := deepseek.DeepSeekChat
	if *conf.Mode == conf.ComplexMode {
		userInfo, err := db.GetUserByName(update.Message.From.String())
		if err != nil {
			log.Printf("Error getting user info: %s\n", err)
		}
		if userInfo != nil && userInfo.Mode != "" {
			log.Printf("User info: %s, %s\n", userInfo.Name, userInfo.Mode)
			model = userInfo.Mode
		}
	}

	client := deepseek.NewClient(*conf.DeepseekToken, *conf.CustomUrl)
	request := &deepseek.StreamChatCompletionRequest{
		Model:  model,
		Stream: true,
	}
	messages := make([]deepseek.ChatCompletionMessage, 0)

	msgRecords := db.GetMsgRecord(update.Message.From.String())
	if msgRecords != nil {
		for _, record := range msgRecords.AQs {
			log.Println("question:", record.Question, "answer:", record.Answer)
			messages = append(messages, deepseek.ChatCompletionMessage{
				Role:    constants.ChatMessageRoleAssistant,
				Content: record.Question,
			})
			messages = append(messages, deepseek.ChatCompletionMessage{
				Role:    constants.ChatMessageRoleUser,
				Content: record.Answer,
			})
		}
	}
	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: prompt,
	})

	request.Messages = messages

	ctx := context.Background()

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
			if len(msgInfoContent.Content) > OneMsgLen {
				messageChan <- msgInfoContent
				msgInfoContent = &param.MsgInfo{
					SendLen:     FirstSendLen,
					FullContent: msgInfoContent.FullContent,
				}
			}

			msgInfoContent.Content += choice.Delta.Content
			msgInfoContent.FullContent += choice.Delta.Content
			if len(msgInfoContent.Content) > msgInfoContent.SendLen {
				messageChan <- msgInfoContent
				msgInfoContent.SendLen += NonFirstSendLen
			}
		}
	}

	messageChan <- msgInfoContent

	return nil
}

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
