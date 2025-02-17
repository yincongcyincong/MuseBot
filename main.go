package main

import (
	"context"
	"errors"
	"fmt"
	_ "fmt"
	deepseek "github.com/cohesion-org/deepseek-go"
	constants "github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io"
	"log"
	"strings"
)

func main() {
	// 替换为你的Telegram Bot Token
	bot, err := tgbotapi.NewBotAPI("7331299978:AAGL2hdQJE5MMvjWlWM-BfsC4mKnPu84XJY")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	for update := range updates {
		// 检查是否有新消息
		if update.Message != nil {

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.Text == "" || !strings.Contains(update.Message.Text, "@Guanwushan_bot") {
				return
			}

			messageChan := make(chan string)

			// 调用DeepSeek API
			go func(update tgbotapi.Update) {
				text := strings.ReplaceAll(update.Message.Text, "@Guanwushan_bot", "")
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

	client := deepseek.NewClient("sk-e2b91d48bfbf4b8a8091642ae3d6ee9a")
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
		log.Fatalf("ChatCompletionStream error: %v", err)
		return err
	}
	defer stream.Close()
	sendMsg := ""
	limit := 30
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
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

	return nil
}
