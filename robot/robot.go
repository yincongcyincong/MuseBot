package robot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/command"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/deepseek"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

// StartListenRobot start listen robot callback
func StartListenRobot() {
	for {

		bot := conf.CreateBot()
		logger.Info("telegramBot Info", "username", bot.Self.UserName)

		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := bot.GetUpdatesChan(u)
		for update := range updates {
			chatId, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
			if !checkUserAllow(update) && !checkGroupAllow(update) {
				chat := utils.GetChat(update)
				logger.Warn("user/group not allow to use this bot", "userID", userId, "chat", chat)
				i18n.SendMsg(chatId, "valid_user_group", bot, nil)
				continue
			}

			if handleCommandAndCallback(update, bot) {
				continue
			}
			// check whether you have new message
			if update.Message != nil {

				if skipThisMsg(update, bot) {
					continue
				}

				if *conf.DeepseekType == param.DeepSeek {
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
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	if checkUserTokenExceed(update, bot) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}

	messageChan := make(chan *param.MsgInfo)

	// request DeepSeek API
	go deepseek.GetContentFromHS(messageChan, update, bot, content)

	// send response message
	go handleUpdate(messageChan, update, bot, content)
}

// requestDeepseekAndResp request deepseek api
func requestDeepseekAndResp(update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	if checkUserTokenExceed(update, bot) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}

	messageChan := make(chan *param.MsgInfo)

	// request DeepSeek API
	go deepseek.GetContentFromDP(messageChan, update, bot, content)

	// send response message
	go handleUpdate(messageChan, update, bot, content)
}

// handleUpdate handle robot msg sending
func handleUpdate(messageChan chan *param.MsgInfo, update tgbotapi.Update, bot *tgbotapi.BotAPI, content string) {
	var msg *param.MsgInfo

	chatId, msgId, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	parseMode := tgbotapi.ModeMarkdown

	tgMsgInfo := tgbotapi.NewMessage(chatId, i18n.GetMessage(*conf.Lang, "thinking", nil))
	tgMsgInfo.ReplyToMessageID = msgId
	firstSendInfo, err := bot.Send(tgMsgInfo)
	if err != nil {
		logger.Warn("Sending first message fail", "err", err)
	}

	for msg = range messageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from deepseek!"
		}
		if firstSendInfo.MessageID != 0 {
			msg.MsgId = firstSendInfo.MessageID
		}

		if msg.MsgId == 0 && firstSendInfo.MessageID == 0 {
			tgMsgInfo = tgbotapi.NewMessage(chatId, msg.Content)
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
					logger.Warn("Error sending message:", "msgID", msgId, "err", err)
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
			_, err = bot.Send(updateMsg)

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
					logger.Warn("Error editing message", "msgID", msgId, "err", err)
				}
			}
			firstSendInfo.MessageID = 0
		}

	}

	// store question and answer into record.
	if msg != nil && msg.FullContent != "" {
		db.InsertMsgRecord(userId, &db.AQ{
			Question: content,
			Answer:   msg.FullContent,
			Token:    msg.Token,
		}, true)
	} else {
		if !utils.CheckMsgIsCallback(update) {
			db.InsertMsgRecord(userId, &db.AQ{
				Question: content,
				Answer:   "",
				Token:    0,
			}, true)
		}
	}

}

func sleepUtilNoLimit(msgId int, err error) bool {
	var apiErr *tgbotapi.Error
	if errors.As(err, &apiErr) && apiErr.Message == "Too Many Requests" {
		waitTime := time.Duration(apiErr.RetryAfter) * time.Second
		logger.Warn("Rate limited. Retrying after", "msgID", msgId, "waitTime", waitTime)
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
	_, _, userID := utils.GetChatIdAndMsgIdAndUserID(update)
	logger.Info("command info", "userID", userID, "cmd", cmd)
	switch cmd {
	case "chat":
		sendChatMessage(update, bot)
	case "mode":
		sendModeConfigurationOptions(bot, update.Message.Chat.ID)
	case "balance":
		showBalanceInfo(update, bot)
	case "state":
		showStateInfo(update, bot)
	case "clear":
		clearAllRecord(update, bot)
	case "retry":
		retryLastQuestion(update, bot)
	case "photo":
		sendImg(update)
	case "video":
		sendVideo(update)
	case "help":
		sendHelpConfigurationOptions(bot, update.Message.Chat.ID)
	default:
		command.ExecuteCustomCommand(cmd)
	}

	if checkAdminUser(update) {
		switch cmd {
		case "addtoken":
			addToken(update, bot)
		}
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
		i18n.SendMsg(chatId, "chat_fail", bot, nil)
		return
	}

	// Reply to the chat content
	if *conf.DeepseekType == param.DeepSeek {
		requestDeepseekAndResp(update, bot, content)
	} else {
		requestHuoshanAndResp(update, bot, content)
	}
}
func retryLastQuestion(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)

	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		if *conf.DeepseekType == param.DeepSeek {
			requestDeepseekAndResp(update, bot, records.AQs[len(records.AQs)-1].Question)
		} else {
			requestHuoshanAndResp(update, bot, records.AQs[len(records.AQs)-1].Question)
		}
	} else {
		i18n.SendMsg(chatId, "last_question_fail", bot, nil)
	}
}

// clearAllRecord clear all record
func clearAllRecord(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	db.DeleteMsgRecord(userId)
	i18n.SendMsg(chatId, "delete_succ", bot, nil)
}

// addToken clear all record
func addToken(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, _ := utils.GetChatIdAndMsgIdAndUserID(update)
	msg := utils.GetMessage(update)
	command := "/addtoken"
	mention := "@" + bot.Self.UserName

	content := strings.ReplaceAll(msg.Text, command, mention)
	content = strings.ReplaceAll(content, mention, "")
	content = strings.TrimSpace(content)
	splitContent := strings.Split(content, " ")

	db.AddAvailToken(int64(utils.ParseInt(splitContent[0])), utils.ParseInt(splitContent[1]))
	i18n.SendMsg(chatId, "add_token_succ", bot, nil)
}

// showBalanceInfo show balance info
func showBalanceInfo(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, _ := utils.GetChatIdAndMsgIdAndUserID(update)

	if *conf.DeepseekType != param.DeepSeek {
		i18n.SendMsg(chatId, "not_deepseek", bot, nil)
		return
	}

	balance := deepseek.GetBalanceInfo()

	// handle balance info msg
	msgContent := fmt.Sprintf(i18n.GetMessage(*conf.Lang, "balance_title", nil), balance.IsAvailable)

	template := i18n.GetMessage(*conf.Lang, "balance_content", nil)

	for _, bInfo := range balance.BalanceInfos {
		msgContent += fmt.Sprintf(template, bInfo.Currency, bInfo.TotalBalance,
			bInfo.ToppedUpBalance, bInfo.GrantedBalance)
	}

	utils.SendMsg(chatId, msgContent, bot)
}

// showStateInfo show user's usage of token
func showStateInfo(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	chatId, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		return
	}

	// get today token
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	todayTokey, err := db.GetTokenByUserIdAndTime(userId, startOfDay.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get today token fail", "err", err)
	}

	// get this week token
	startOf7DaysAgo := now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
	weekToken, err := db.GetTokenByUserIdAndTime(userId, startOf7DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get week token fail", "err", err)
	}

	// handle balance info msg
	startOf30DaysAgo := now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	monthToken, err := db.GetTokenByUserIdAndTime(userId, startOf30DaysAgo.Unix(), endOfDay.Unix())
	if err != nil {
		logger.Warn("get week token fail", "err", err)
	}

	template := i18n.GetMessage(*conf.Lang, "state_content", nil)
	msgContent := fmt.Sprintf(template, userInfo.Token, todayTokey, weekToken, monthToken)
	utils.SendMsg(chatId, msgContent, bot)
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

	i18n.SendMsg(chatID, "chat_mode", bot, &inlineKeyboard)
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

	i18n.SendMsg(chatID, "command_notice", bot, &inlineKeyboard)
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
		logger.Warn("get user fail", "userID", update.CallbackQuery.From.ID, "err", err)
		sendFailMessage(update, bot)
		return
	}

	if userInfo != nil && userInfo.ID != 0 {
		err = db.UpdateUserMode(update.CallbackQuery.From.ID, update.CallbackQuery.Data)
		if err != nil {
			logger.Warn("update user fail", "userID", update.CallbackQuery.From.ID, "err", err)
			sendFailMessage(update, bot)
			return
		}
	} else {
		_, err = db.InsertUser(update.CallbackQuery.From.ID, update.CallbackQuery.Data)
		if err != nil {
			logger.Warn("insert user fail", "userID", update.CallbackQuery.From.String(), "err", err)
			sendFailMessage(update, bot)
			return
		}
	}

	// send response
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		logger.Warn("request callback fail", "err", err)
	}

	utils.SendMsg(update.CallbackQuery.Message.Chat.ID,
		i18n.GetMessage(*conf.Lang, "mode_choose", nil)+update.CallbackQuery.Data, bot)
}

func sendFailMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, i18n.GetMessage(*conf.Lang, "set_mode", nil))
	if _, err := bot.Request(callback); err != nil {
		logger.Warn("request callback fail", "err", err)
	}

	i18n.SendMsg(update.CallbackQuery.Message.Chat.ID, "set_mode", bot, nil)
}

func sendVideo(update tgbotapi.Update) {
	prompt := strings.Replace(update.Message.Text, "/video", "", 1)
	videoUrl, err := deepseek.GenerateVideo(prompt)
	if err != nil {
		logger.Warn("generate video fail", "err", err)
		return
	}

	if len(videoUrl) == 0 {
		logger.Warn("no video generated")
		return
	}

	// create image url
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendVideo", *conf.BotToken)
	chatId, replyToMessageID, userId := utils.GetChatIdAndMsgIdAndUserID(update)

	// construct request param
	req := map[string]interface{}{
		"chat_id": chatId,
		"video":   videoUrl,
	}
	if replyToMessageID != 0 {
		req["reply_to_message_id"] = replyToMessageID
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Warn("marshal json content fail", "err", err)
		return
	}

	// send post request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Warn("send request fail", "err", err)
		return
	}
	defer resp.Body.Close()

	// analysis response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Warn("analysis response fail", "err", err)
		return
	}

	if ok, found := result["ok"].(bool); !found || !ok {
		logger.Warn("send video fail", "result", result)
		return
	}

	db.AddToken(userId, param.VideoTokenUsage)
	return
}

func sendImg(update tgbotapi.Update) {
	prompt := strings.Replace(update.Message.Text, "/photo", "", 1)
	data, err := deepseek.GenerateImg(prompt)
	if err != nil {
		logger.Warn("generate image fail", "err", err)
		return
	}

	if data.Data == nil || len(data.Data.ImageUrls) == 0 {
		logger.Warn("no image generated")
		return
	}

	// create image url
	photoURL := data.Data.ImageUrls[0]
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", *conf.BotToken)
	chatId, replyToMessageID, userId := utils.GetChatIdAndMsgIdAndUserID(update)

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
		logger.Warn("marshal json content fail", "err", err)
		return
	}

	// send post request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Warn("send request fail", "err", err)
		return
	}
	defer resp.Body.Close()

	// analysis response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Warn("analysis response fail", "err", err)
		return
	}

	if ok, found := result["ok"].(bool); !found || !ok {
		logger.Warn("send image fail", "result", result)
		return
	}

	db.AddToken(userId, param.ImageTokenUsage)

	return
}

func checkUserAllow(update tgbotapi.Update) bool {
	if len(conf.AllowedTelegramUserIds) == 0 {
		return true
	}
	if conf.AllowedTelegramUserIds[0] {
		return false
	}

	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	_, ok := conf.AllowedTelegramUserIds[userId]
	return ok
}

func checkGroupAllow(update tgbotapi.Update) bool {
	chat := utils.GetChat(update)
	if chat == nil {
		return false
	}

	if chat.IsGroup() || chat.IsSuperGroup() { // 判断是否是群组或超级群组
		if len(conf.AllowedTelegramGroupIds) == 0 {
			return true
		}
		if conf.AllowedTelegramGroupIds[0] {
			return false
		}
		if _, ok := conf.AllowedTelegramGroupIds[chat.ID]; ok {
			return true
		}
	}

	return false
}

func checkUserTokenExceed(update tgbotapi.Update, bot *tgbotapi.BotAPI) bool {
	if *conf.TokenPerUser == 0 {
		return false
	}

	chatId, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		return false
	}

	if userInfo == nil {
		logger.Warn("get user info is nil")
		return false
	}

	if userInfo.AvailToken >= userInfo.Token {
		tpl := i18n.GetMessage(*conf.Lang, "token_exceed", nil)
		content := fmt.Sprintf(tpl, userInfo.Token, userInfo.AvailToken-userInfo.Token)
		utils.SendMsg(chatId, content, bot)
		return true
	}

	return false
}

func checkAdminUser(update tgbotapi.Update) bool {
	if len(conf.AdminUserIds) == 0 {
		return false
	}

	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(update)
	_, ok := conf.AdminUserIds[userId]
	return ok
}
