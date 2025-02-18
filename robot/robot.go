package robot

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/deepseek"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

func StartListenRobot() {
	// 替换为你的Telegram Bot Token
	bot, err := tgbotapi.NewBotAPI(*conf.BotToken)
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

			messageChan := make(chan *param.MsgInfo)

			// request DeepSeek API
			go deepseek.GetContentFromDP(messageChan, update, bot)

			// send response message
			go handleUpdate(messageChan, update, bot)

		}
	}
}

func handleUpdate(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	for msg := range messageChan {
		if len(msg.Content) == 0 {
			log.Printf("%d content len is 0\n", update.Message.MessageID)
			continue
		}

		if msg.MsgId == 0 {
			tgMsgInfo := tgbotapi.NewMessage(update.Message.Chat.ID, msg.Content)
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
			msg.MsgId = sendInfo.MessageID
		} else {
			updateMsg := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:    update.Message.Chat.ID,
					MessageID: msg.MsgId,
				},
				Text: msg.Content,
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
