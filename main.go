package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	deepseek "github.com/cohesion-org/deepseek-go"
	constants "github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"os"
	"strings"
)

var (
	BotToken      *string
	DeepseekToken *string
)

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
		// 检查是否有新消息
		if update.Message != nil {

			fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

			if update.Message.Text == "" || !strings.Contains(update.Message.Text, "@"+bot.Self.UserName) {
				continue
			}

			messageChan := make(chan string)

			// 调用DeepSeek API
			go func(update tgbotapi.Update) {
				text := strings.ReplaceAll(update.Message.Text, "@"+bot.Self.UserName, "")
				err := callDeepSeekAPI(text, messageChan)
				if err != nil {
					log.Printf("Error calling DeepSeek API: %s", err)
				}
				close(messageChan)
			}(update)

			// 发送回复消息
			go func(update tgbotapi.Update) {
				msgId := 0
				for msg := range messageChan {
					if msgId == 0 {
						msgInfo := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
						msgInfo.ReplyToMessageID = update.Message.MessageID
						sendInfo, err := bot.Send(msgInfo)
						if err != nil {
							log.Printf("Error sending message: %s", err)
						}
						msgId = sendInfo.MessageID
					} else {
						_, err := bot.Send(tgbotapi.EditMessageTextConfig{
							BaseEdit: tgbotapi.BaseEdit{
								ChatID:    update.Message.Chat.ID,
								MessageID: msgId,
							},
							Text: msg,
						})
						if err != nil {
							log.Println("Error editing message:", err)
						}
					}

				}

			}(update)

		}
	}
}

// callDeepSeekAPI 调用DeepSeek API并返回响应
func callDeepSeekAPI(prompt string, messageChan chan string) error {

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
		log.Printf("ChatCompletionStream error: %v\n", err)
		return err
	}
	defer stream.Close()
	sendMsg := ""
	limit := 30
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Printf("\nStream finished")
			break
		}
		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			break
		}
		for _, choice := range response.Choices {
			sendMsg += choice.Delta.Content
			if len(sendMsg) > limit {
				messageChan <- sendMsg
				limit += 100
			}
		}
	}

	messageChan <- sendMsg

	return nil
}
