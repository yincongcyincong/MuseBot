package robot

import (
	"errors"
	"fmt"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
	"log"
	"strings"
	"time"

	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/deepseek"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

func StartListenRobot() {
	for {
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
			if handleCommandAndCallback(update, bot) {
				continue
			}
			// check whether you have new message
			if update.Message != nil {

				if skipThisMsg(update, bot) {
					continue
				}

				if *conf.DeepseekType == "deepseek" {
					requestDeepseekAndResp(update, bot, update.Message.Text)
				} else {
					requestHuoshanAndResp(update, bot, update.Message.Text)
				}

			}
		}
	}
}

func requestHuoshanAndResp(update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	messageChan := make(chan *param.MsgInfo)

	// request DeepSeek API
	go deepseek.GetContentFromHS(messageChan, update, bot, content)

	// send response message
	go handleUpdate(messageChan, update, bot, content)
}

func requestDeepseekAndResp(update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	messageChan := make(chan *param.MsgInfo)

	// request DeepSeek API
	go deepseek.GetContentFromDP(messageChan, update, bot, content)

	// send response message
	go handleUpdate(messageChan, update, bot, content)
}

func handleUpdate(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	var msg *param.MsgInfo

	chatId, msgId, username := utils.GetChatIdAndMsgIdAndUserName(update)
	for msg = range messageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from deepseek!"
		}

		if msg.MsgId == 0 {
			tgMsgInfo := tgbotapi.NewMessage(chatId, msg.Content)
			tgMsgInfo.ReplyToMessageID = msgId
			tgMsgInfo.ParseMode = tgbotapi.ModeMarkdown
			sendInfo, err := bot.Send(tgMsgInfo)
			if err != nil {
				if sleepUtilNoLimit(msgId, err) {
					sendInfo, err = bot.Send(tgMsgInfo)
				} else {
					sendInfo, err = bot.Send(tgMsgInfo)
				}
				if err != nil {
					log.Printf("%d Error sending message: %s\n", msgId, err)
					continue
				}
			}
			msg.MsgId = sendInfo.MessageID
		} else {
			updateMsg := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:    chatId,
					MessageID: msg.MsgId,
				},
				Text:      msg.Content,
				ParseMode: tgbotapi.ModeMarkdown,
			}
			_, err := bot.Send(updateMsg)

			if err != nil {
				// try again
				if sleepUtilNoLimit(msgId, err) {
					_, err = bot.Send(updateMsg)
				} else {
					_, err = bot.Send(updateMsg)
				}
				if err != nil {
					log.Printf("Error editing message:%d %s\n", msgId, err)
				}
			}
		}

	}

	// store question and answer into record.
	if msg != nil && msg.FullContent != "" {
		db.InsertMsgRecord(username, &db.AQ{
			Question: content,
			Answer:   msg.FullContent,
		}, true)
	} else {
		if !utils.CheckMsgIsCallback(update) {
			db.InsertMsgRecord(username, &db.AQ{
				Question: content,
				Answer:   "",
			}, true)
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

func handleCommandAndCallback(update tgbotapi.Update, bot *tgbotapi.BotAPI) bool {
	// if it's command, directly
	if update.Message != nil && update.Message.IsCommand() {
		handleCommand(update, bot)
		return true
	}

	if update.CallbackQuery != nil {
		handleCallbackQuery(update, bot)
		return true
	}
	return false
}

func skipThisMsg(update tgbotapi.Update, bot *tgbotapi.BotAPI) bool {

	if update.Message.Chat.Type == "private" {
		return false
	}

	if update.Message.Text == "" || !strings.Contains(update.Message.Text, "@"+bot.Self.UserName) {
		return true
	}

	return false
}

func handleCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	cmd := update.Message.Command()
	switch cmd {
	case "mode":
		sendModeConfigurationOptions(bot, update.Message.Chat.ID)
	case "balance":
		showBalanceInfo(update, bot)
	case "clear":
		clearAllRecord(update, bot)
	case "retry":
		retryLastQuestion(update, bot)
	case "help":
		sendHelpConfigurationOptions(bot, update.Message.Chat.ID)
	}
}

func retryLastQuestion(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, username := utils.GetChatIdAndMsgIdAndUserName(update)

	records := db.GetMsgRecord(username)
	if records != nil && len(records.AQs) > 0 {
		requestDeepseekAndResp(update, bot, records.AQs[len(records.AQs)-1].Question)
	} else {
		msg := tgbotapi.NewMessage(chatId, "🚀no last question!")
		msg.ParseMode = tgbotapi.ModeMarkdown
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("send retry message fail: %v\n", err)
		}
	}
}

func clearAllRecord(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, username := utils.GetChatIdAndMsgIdAndUserName(update)
	db.DeleteMsgRecord(username)
	msg := tgbotapi.NewMessage(chatId, "🚀successfully delete!")
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send clear message fail: %v\n", err)
	}
}

func showBalanceInfo(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, _ := utils.GetChatIdAndMsgIdAndUserName(update)

	if *conf.DeepseekType != "deepseek" {
		msg := tgbotapi.NewMessage(chatId, "🚀no-deepseek")
		msg.ParseMode = tgbotapi.ModeMarkdown
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("send message fail: %v\n", err)
		}
		return
	}

	balance := deepseek.GetBalanceInfo()

	// handle balance info msg
	msgContent := fmt.Sprintf(`🟣 Available: %t

`, balance.IsAvailable)

	template := `🟣 Your Currency: %s

🟣 Your TotalBalance Left: %s

🟣 Your ToppedUpBalance Left: %s

🟣 Your GrantedBalance Left: %s

`
	for _, bInfo := range balance.BalanceInfos {
		msgContent += fmt.Sprintf(template, bInfo.Currency, bInfo.TotalBalance,
			bInfo.ToppedUpBalance, bInfo.GrantedBalance)
	}

	msg := tgbotapi.NewMessage(chatId, msgContent)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send balance message fail: %v\n", err)
	}

}

// 发送配置选择界面
func sendModeConfigurationOptions(bot *tgbotapi.BotAPI, chatID int64) {
	// create inline button
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("chat", godeepseek.DeepSeekChat),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("coder", godeepseek.DeepSeekCoder),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("reasoner", godeepseek.DeepSeekReasoner),
		),
	)

	// 发送消息并附上内联键盘
	msg := tgbotapi.NewMessage(chatID, "🚀**Select chat mode**")
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send inline message fail: %v\n", err)
	}
}

func sendHelpConfigurationOptions(bot *tgbotapi.BotAPI, chatID int64) {
	// create inline button
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("mode", "mode"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("balance", "balance"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("retry", "retry"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("clear", "clear"),
		),
	)

	// 发送消息并附上内联键盘
	msg := tgbotapi.NewMessage(chatID, "🤖**Select command**")
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send inline message fail: %v\n", err)
	}
}

// 处理回调查询（用户点击按钮）
func handleCallbackQuery(update tgbotapi.Update, bot *tgbotapi.BotAPI) {

	switch update.CallbackQuery.Data {
	case godeepseek.DeepSeekChat, godeepseek.DeepSeekCoder, godeepseek.DeepSeekReasoner:
		handleModeUpdate(update, bot)
	case "mode":
		sendModeConfigurationOptions(bot, update.CallbackQuery.Message.Chat.ID)
	case "balance":
		showBalanceInfo(update, bot)
	case "clear":
		clearAllRecord(update, bot)
	case "retry":
		retryLastQuestion(update, bot)
	}

}

func handleModeUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userInfo, err := db.GetUserByName(update.CallbackQuery.From.String())
	if err != nil {
		log.Printf("get user fail: %s %v", update.CallbackQuery.From.String(), err)
		sendFailMessage(update, bot)
		return
	}

	if userInfo != nil && userInfo.ID != 0 {
		err = db.UpdateUserMode(update.CallbackQuery.From.String(), update.CallbackQuery.Data)
		if err != nil {
			log.Printf("update user fail: %s %v\n", update.CallbackQuery.From.String(), err)
			sendFailMessage(update, bot)
			return
		}
	} else {
		_, err = db.InsertUser(update.CallbackQuery.From.String(), update.CallbackQuery.Data)
		if err != nil {
			log.Printf("insert user fail: %s %v\n", update.CallbackQuery.From.String(), err)
			sendFailMessage(update, bot)
			return
		}
	}

	// send response
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		log.Printf("request callback fail: %v\n", err)
	}

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "You choose: "+update.CallbackQuery.Data)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("request send msg fail: %v\n", err)
	}
}

func sendFailMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "set mode fail!")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("request callback fail: %v\n", err)
	}

	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "set mode fail!")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("request send msg fail: %v\n", err)
	}
}
