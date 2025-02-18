package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
	constants "github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	OneMsgLen       = 1500
	FirstSendLen    = 30
	NonFirstSendLen = 300
)

var (
	BotToken      *string
	DeepseekToken *string
)

type msgInfo struct {
	msgId   int
	content string
	sendLen int
}

func main() {
	BotToken = flag.String("telegram_bot_token", "", "Comma-separated list of Telegram bot tokens")
	DeepseekToken = flag.String("deepseek_token", "", "deepseek auth token")
	flag.Parse()

	if *BotToken == "" {
		*BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	if *DeepseekToken == "" {
		*DeepseekToken = os.Getenv("DEEPSEEK_TOKEN")
	}

	fmt.Println("TelegramBotToken:", *BotToken)
	fmt.Println("DeepseekToken:", *DeepseekToken)
	if *BotToken == "" || *DeepseekToken == "" {
		log.Fatalf("Bot token and deepseek token are required")
	}

	// 替换为你的Telegram Bot Token
	bot, err := tgbotapi.NewBotAPI(*BotToken)
	if err != nil {
		log.Fatalf("Init bot fail: %v\n", err.Error())
	}

	bot.Debug = true

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		// check whether you have new message
		if update.Message != nil {

			fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

			if update.Message.Text == "" || !strings.Contains(update.Message.Text, "@"+bot.Self.UserName) {
				continue
			}

			messageChan := make(chan *msgInfo)

			// request DeepSeek API
			go func(update tgbotapi.Update) {
				text := strings.ReplaceAll(update.Message.Text, "@"+bot.Self.UserName, "")
				err := callDeepSeekAPI(text, update.Message.MessageID, messageChan)
				if err != nil {
					log.Printf("Error calling DeepSeek API: %s\n", err)
				}
				close(messageChan)
			}(update)

			// send response message
			go func(update tgbotapi.Update) {
				for msg := range messageChan {
					if len(msg.content) == 0 {
						log.Printf("%d content len is 0\n", update.Message.MessageID)
						continue
					}

					if msg.msgId == 0 {
						tgMsgInfo := tgbotapi.NewMessage(update.Message.Chat.ID, msg.content)
						tgMsgInfo.ReplyToMessageID = update.Message.MessageID
						sendInfo, err := bot.Send(tgMsgInfo)
						if err != nil {
							if sleepUtilNoLimit(update.Message.MessageID, err) {
								sendInfo, err = bot.Send(tgMsgInfo)
							}
							if err != nil {
								log.Printf("%d Error sending message: %s\n", update.Message.MessageID, err)
								continue
							}
						}
						msg.msgId = sendInfo.MessageID
					} else {
						updateMsg := tgbotapi.EditMessageTextConfig{
							BaseEdit: tgbotapi.BaseEdit{
								ChatID:    update.Message.Chat.ID,
								MessageID: msg.msgId,
							},
							Text: msg.content,
						}
						_, err := bot.Send(updateMsg)

						if err != nil {
							// try again
							if sleepUtilNoLimit(update.Message.MessageID, err) {
								_, err = bot.Send(updateMsg)
							}
							if err != nil {
								log.Printf("Error editing message:%d %s\n", update.Message.MessageID, err)
							}
						}
					}

				}

			}(update)

		}
	}
}

// callDeepSeekAPI request DeepSeek API and get response
func callDeepSeekAPI(prompt string, updateMsgID int, messageChan chan *msgInfo) error {

	client := deepseek.NewClient(*DeepseekToken)
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
	msgInfoContent := &msgInfo{
		sendLen: FirstSendLen,
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
			if len(msgInfoContent.content) > OneMsgLen {
				messageChan <- msgInfoContent
				msgInfoContent = &msgInfo{
					sendLen: FirstSendLen,
				}
			}

			msgInfoContent.content += choice.Delta.Content
			if len(msgInfoContent.content) > msgInfoContent.sendLen {
				messageChan <- msgInfoContent
				msgInfoContent.sendLen += NonFirstSendLen
			}
		}
	}

	messageChan <- msgInfoContent

	return nil
}

func sleepUtilNoLimit(msgId int, err error) bool {
	var apiErr *tgbotapi.Error
	if errors.As(err, &apiErr) && apiErr.Message == "Too Many Requests" {
		waitTime := time.Duration(apiErr.RetryAfter) * time.Second
		fmt.Printf("Rate limited. Retrying after %d %v...\n", msgId, waitTime)
		time.Sleep(waitTime)
		return true
	}

	return false
}
