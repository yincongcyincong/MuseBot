package robot

import (
	"errors"
	"fmt"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
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

			fmt.Printf("[%s] %s\n", update.Message.From.String(), update.Message.Text)

			if skipThisMsg(update, bot) {
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
	var msg *param.MsgInfo
	for msg = range messageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from deepseek!"
		}

		if msg.MsgId == 0 {
			tgMsgInfo := tgbotapi.NewMessage(update.Message.Chat.ID, msg.Content)
			tgMsgInfo.ReplyToMessageID = update.Message.MessageID
			tgMsgInfo.ParseMode = tgbotapi.ModeMarkdown
			sendInfo, err := bot.Send(tgMsgInfo)
			if err != nil {
				if sleepUtilNoLimit(update.Message.MessageID, err) {
					sendInfo, err = bot.Send(tgMsgInfo)
				} else {
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
				Text:      msg.Content,
				ParseMode: tgbotapi.ModeMarkdown,
			}
			_, err := bot.Send(updateMsg)

			if err != nil {
				// try again
				if sleepUtilNoLimit(update.Message.MessageID, err) {
					_, err = bot.Send(updateMsg)
				} else {
					_, err = bot.Send(updateMsg)
				}
				if err != nil {
					log.Printf("Error editing message:%d %s\n", update.Message.MessageID, err)
				}
			}
		}

	}

	// store question and answer into record.
	if msg != nil && msg.FullContent != "" {
		db.InsertMsgRecord(update.Message.From.String(), &db.AQ{
			Question: update.Message.Text,
			Answer:   msg.FullContent,
		})
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
	if update.Message != nil && update.Message.IsCommand() && *conf.Mode == conf.ComplexMode {
		handleCommand(update, bot)
		return true
	}

	if update.CallbackQuery != nil && *conf.Mode == conf.ComplexMode {
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
	case "help":
		sendHelpConfigurationOptions(bot, update.Message.Chat.ID)
	}
}

func clearAllRecord(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	db.DeleteMsgRecord(update.Message.From.String())
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚀successfully delete")
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send clear message fail: %v\n", err)
	}
}

func showBalanceInfo(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
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

	chatId := int64(0)
	if update.Message != nil {
		chatId = update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		chatId = update.CallbackQuery.Message.Chat.ID
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
			//tgbotapi.NewInlineKeyboardButtonData("retry", "retry"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("balance", "balance"),
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
