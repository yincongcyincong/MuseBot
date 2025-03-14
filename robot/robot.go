package robot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/deepseek"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

// StartListenRobot start listen robot callback
func StartListenRobot() {
	for {
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

// requestHuoshanAndResp request huoshan api
func requestHuoshanAndResp(update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	messageChan := make(chan *param.MsgInfo)

	// request DeepSeek API
	go deepseek.GetContentFromHS(messageChan, update, bot, content)

	// send response message
	go handleUpdate(messageChan, update, bot, content)
}

// requestDeepseekAndResp request deepseek api
func requestDeepseekAndResp(update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	messageChan := make(chan *param.MsgInfo)

	// request DeepSeek API
	go deepseek.GetContentFromDP(messageChan, update, bot, content)

	// send response message
	go handleUpdate(messageChan, update, bot, content)
}

// handleUpdate handle robot msg sending
func handleUpdate(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	var msg *param.MsgInfo

	chatId, msgId, username := utils.GetChatIdAndMsgIdAndUserID(update)
	parseMode := tgbotapi.ModeMarkdown

	for msg = range messageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from deepseek!"
		}

		if msg.MsgId == 0 {
			tgMsgInfo := tgbotapi.NewMessage(chatId, msg.Content)
			tgMsgInfo.ReplyToMessageID = msgId
			tgMsgInfo.ParseMode = parseMode
			sendInfo, err := bot.Send(tgMsgInfo)
			if err != nil {
				if sleepUtilNoLimit(msgId, err) {
					sendInfo, err = bot.Send(tgMsgInfo)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					tgMsgInfo.ParseMode = ""
					sendInfo, err = bot.Send(tgMsgInfo)
				} else {
					_, err = bot.Send(tgMsgInfo)
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
				ParseMode: parseMode,
			}
			_, err := bot.Send(updateMsg)

			if err != nil {
				// try again
				if sleepUtilNoLimit(msgId, err) {
					_, err = bot.Send(updateMsg)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					updateMsg.ParseMode = ""
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
	case "chat":
		sendChatMessage(update, bot)
	case "mode":
		sendModeConfigurationOptions(bot, update.Message.Chat.ID)
	case "balance":
		showBalanceInfo(update, bot)
	case "clear":
		clearAllRecord(update, bot)
	case "retry":
		retryLastQuestion(update, bot)
	case "photo":
		sendImg(update)
	case "help":
		sendHelpConfigurationOptions(bot, update.Message.Chat.ID)
	}
}

func sendChatMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, _ := utils.GetChatIdAndMsgIdAndUserID(update)
	messageText := update.Message.Text

	// Remove /chat and /chat@botUserName from the message
	command := "/chat"
	mention := "@" + bot.Self.UserName

	content := strings.ReplaceAll(messageText, command, mention)
	content = strings.ReplaceAll(content, mention, "")
	content = strings.TrimSpace(content)

	if len(content) == 0 {
		// If there is no chat content after command
		msg := tgbotapi.NewMessage(chatId, "❌ Please input text after /chat command.")
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("send help message fail: %v\n", err)
		}
		return
	}

	// Reply to the chat content
	if *conf.DeepseekType == "deepseek" {
		requestDeepseekAndResp(update, bot, content)
	} else {
		requestHuoshanAndResp(update, bot, content)
	}
}
func retryLastQuestion(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)

	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		if *conf.DeepseekType == "deepseek" {
			requestDeepseekAndResp(update, bot, records.AQs[len(records.AQs)-1].Question)
		} else {
			requestHuoshanAndResp(update, bot, records.AQs[len(records.AQs)-1].Question)
		}
	} else {
		msg := tgbotapi.NewMessage(chatId, "🚀no last question!")
		msg.ParseMode = tgbotapi.ModeMarkdown
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("send retry message fail: %v\n", err)
		}
	}
}

// clearAllRecord clear all record
func clearAllRecord(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	db.DeleteMsgRecord(userId)
	msg := tgbotapi.NewMessage(chatId, "🚀successfully delete!")
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send clear message fail: %v\n", err)
	}
}

// showBalanceInfo show balance info
func showBalanceInfo(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, _ := utils.GetChatIdAndMsgIdAndUserID(update)

	if *conf.DeepseekType != "deepseek" {
		msg := tgbotapi.NewMessage(chatId, "🚀now model is not deepseek")
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

// sendModeConfigurationOptions send config view
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

	msg := tgbotapi.NewMessage(chatID, "🤖**Select command**")
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("send inline message fail: %v\n", err)
	}
}

// handleCallbackQuery handle callback response
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

// handleModeUpdate handle mode update
func handleModeUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userInfo, err := db.GetUserByID(update.CallbackQuery.From.ID)
	if err != nil {
		log.Printf("get user fail: %d %v", update.CallbackQuery.From.ID, err)
		sendFailMessage(update, bot)
		return
	}

	if userInfo != nil && userInfo.ID != 0 {
		err = db.UpdateUserMode(update.CallbackQuery.From.ID, update.CallbackQuery.Data)
		if err != nil {
			log.Printf("update user fail: %d %v\n", update.CallbackQuery.From.ID, err)
			sendFailMessage(update, bot)
			return
		}
	} else {
		_, err = db.InsertUser(update.CallbackQuery.From.ID, update.CallbackQuery.Data)
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

func sendImg(update tgbotapi.Update) {
	prompt := strings.Replace(update.Message.Text, "/photo", "", 1)
	data, err := deepseek.GenerateImg(prompt)
	if err != nil {
		log.Printf("generate image fail: %v\n", err)
		return
	}

	if data.Data == nil || len(data.Data.ImageUrls) == 0 {
		log.Println("no image generated")
		return
	}

	// create image url
	photoURL := data.Data.ImageUrls[0]
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", *conf.BotToken)
	chatId, replyToMessageID, _ := utils.GetChatIdAndMsgIdAndUserID(update)

	// construct request param
	req := map[string]interface{}{
		"chat_id": chatId,
		"photo":   photoURL,
	}
	if replyToMessageID != 0 {
		req["reply_to_message_id"] = replyToMessageID
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("marshal json content fail: %s\n", err)
		return
	}

	// send post request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("send request fail: %s\n", err)
		return
	}
	defer resp.Body.Close()

	// analysis response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("analysis response fail: %s\n", err)
		return
	}

	if ok, found := result["ok"].(bool); !found || !ok {
		log.Printf("send image fail: %+v", result)
		return
	}

	return
}
