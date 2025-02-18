package deepseek

import (
	"context"
	"errors"
	"fmt"
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
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
	err := callDeepSeekAPI(text, update.Message.MessageID, messageChan)
	if err != nil {
		log.Printf("Error calling DeepSeek API: %s\n", err)
	}
	close(messageChan)
}

// callDeepSeekAPI request DeepSeek API and get response
func callDeepSeekAPI(prompt string, updateMsgID int, messageChan chan *param.MsgInfo) error {

	client := deepseek.NewClient(*conf.DeepseekToken)
	request := &deepseek.StreamChatCompletionRequest{
		Model: deepseek.DeepSeekChat,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleUser, Content: prompt},
		},
		Stream: true,
	}
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
					SendLen: FirstSendLen,
				}
			}

			msgInfoContent.Content += choice.Delta.Content
			if len(msgInfoContent.Content) > msgInfoContent.SendLen {
				messageChan <- msgInfoContent
				msgInfoContent.SendLen += NonFirstSendLen
			}
		}
	}

	messageChan <- msgInfoContent

	return nil
}
