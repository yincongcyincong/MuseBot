package robot

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/rag"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type TelegramRobot struct {
	update tgbotapi.Update
	bot    *tgbotapi.BotAPI
}

func NewTelegramRobot(update tgbotapi.Update, bot *tgbotapi.BotAPI) *TelegramRobot {
	return &TelegramRobot{
		update: update,
		bot:    bot,
	}
}

// StartTelegramRobot start listen robot callback
func StartTelegramRobot() {
	for {
		bot := utils.CreateBot()
		logger.Info("telegramBot Info", "username", bot.Self.UserName)
		
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		
		updates := bot.GetUpdatesChan(u)
		for update := range updates {
			t := NewTelegramRobot(update, bot)
			t.execUpdate()
		}
	}
}

// execUpdate exec telegram message
func (t *TelegramRobot) execUpdate() {
	chatId, msgId, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	
	if !t.checkUserAllow() && !t.checkGroupAllow() {
		chat := utils.GetChat(t.update)
		logger.Warn("user/group not allow to use this bot", "userID", userId, "chat", chat)
		i18n.SendMsg(chatId, "valid_user_group", t.bot, nil, msgId)
		return
	}
	
	if t.handleCommandAndCallback() {
		return
	}
	// check whether you have new message
	if t.update.Message != nil {
		if t.skipThisMsg() {
			logger.Warn("skip this msg", "msgId", msgId, "chat", chatId, "type", t.update.Message.Chat.Type, "content", t.update.Message.Text)
			return
		}
		t.requestDeepseekAndResp(t.update.Message.Text)
	}
	
}

// requestDeepseekAndResp request deepseek api
func (t *TelegramRobot) requestDeepseekAndResp(content string) {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	if t.checkUserTokenExceed() {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	if conf.RagConfInfo.Store != nil {
		t.executeChain(content)
	} else {
		t.executeLLM(content)
	}
	
}

// executeChain use langchain to interact llm
func (t *TelegramRobot) executeChain(content string) {
	messageChan := make(chan *param.MsgInfo)
	
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("GetContent panic err", "err", err)
			}
			utils.DecreaseUserChat(t.update)
			close(messageChan)
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		text, err := utils.GetContent(t.update, t.bot, content)
		if err != nil {
			logger.Error("get content fail", "err", err)
			return
		}
		
		dpLLM := rag.NewRag(llm.WithBot(t.bot), llm.WithUpdate(t.update),
			llm.WithMessageChan(messageChan), llm.WithContent(content))
		
		qaChain := chains.NewRetrievalQAFromLLM(
			dpLLM,
			vectorstores.ToRetriever(conf.RagConfInfo.Store, 3),
		)
		_, err = chains.Run(ctx, qaChain, text)
		if err != nil {
			logger.Warn("execute chain fail", "err", err)
		}
	}()
	
	// send response message
	go t.handleUpdate(messageChan)
	
}

// executeLLM directly interact llm
func (t *TelegramRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	l := llm.NewLLM(llm.WithBot(t.bot), llm.WithUpdate(t.update),
		llm.WithMessageChan(messageChan), llm.WithContent(content),
		llm.WithTaskTools(&conf.AgentInfo{
			DeepseekTool:    conf.DeepseekTools,
			VolTool:         conf.VolTools,
			OpenAITools:     conf.OpenAITools,
			GeminiTools:     conf.GeminiTools,
			OpenRouterTools: conf.OpenRouterTools,
		}))
	
	// request DeepSeek API
	go l.GetContent()
	
	// send response message
	go t.handleUpdate(messageChan)
	
}

// handleUpdate handle robot msg sending
func (t *TelegramRobot) handleUpdate(messageChan chan *param.MsgInfo) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	var msg *param.MsgInfo
	
	chatId, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(t.update)
	parseMode := tgbotapi.ModeMarkdown
	
	tgMsgInfo := tgbotapi.NewMessage(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil))
	tgMsgInfo.ReplyToMessageID = msgId
	firstSendInfo, err := t.bot.Send(tgMsgInfo)
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
			sendInfo, err := t.bot.Send(tgMsgInfo)
			if err != nil {
				if sleepUtilNoLimit(msgId, err) {
					sendInfo, err = t.bot.Send(tgMsgInfo)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					tgMsgInfo.ParseMode = ""
					sendInfo, err = t.bot.Send(tgMsgInfo)
				} else {
					_, err = t.bot.Send(tgMsgInfo)
				}
				if err != nil {
					logger.Warn("Error sending message:", "msgID", msgId, "err", err)
					continue
				}
			}
			msg.MsgId = sendInfo.MessageID
		} else {
			updateMsg := tgbotapi.NewEditMessageText(chatId, msg.MsgId, msg.Content)
			updateMsg.ParseMode = parseMode
			_, err = t.bot.Send(updateMsg)
			if err != nil {
				// try again
				if sleepUtilNoLimit(msgId, err) {
					_, err = t.bot.Send(updateMsg)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					updateMsg.ParseMode = ""
					_, err = t.bot.Send(updateMsg)
				} else {
					_, err = t.bot.Send(updateMsg)
				}
				if err != nil {
					logger.Warn("Error editing message", "msgID", msgId, "err", err)
				}
			}
			firstSendInfo.MessageID = 0
		}
		
	}
}

// sleepUtilNoLimit handle "Too Many Requests" error
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

// handleCommandAndCallback telegram command and callback function
func (t *TelegramRobot) handleCommandAndCallback() bool {
	// if it's command, directly
	if t.update.Message != nil && t.update.Message.IsCommand() {
		go t.handleCommand()
		return true
	}
	
	if t.update.CallbackQuery != nil {
		go t.handleCallbackQuery()
		return true
	}
	
	if t.update.Message != nil && t.update.Message.ReplyToMessage != nil && t.update.Message.ReplyToMessage.From != nil &&
		t.update.Message.ReplyToMessage.From.UserName == t.bot.Self.UserName {
		go t.ExecuteForceReply()
		return true
	}
	
	return false
}

// skipThisMsg check if msg trigger llm
func (t *TelegramRobot) skipThisMsg() bool {
	if t.update.Message.Chat.Type == "private" {
		if strings.TrimSpace(t.update.Message.Text) == "" &&
			t.update.Message.Voice == nil && t.update.Message.Photo == nil {
			return true
		}
		
		return false
	} else {
		if strings.TrimSpace(strings.ReplaceAll(t.update.Message.Text, "@"+t.bot.Self.UserName, "")) == "" &&
			t.update.Message.Voice == nil {
			return true
		}
		
		if !strings.Contains(t.update.Message.Text, "@"+t.bot.Self.UserName) {
			return true
		}
	}
	
	return false
}

// handleCommand handle multiple commands
func (t *TelegramRobot) handleCommand() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleCommand panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	cmd := t.update.Message.Command()
	_, _, userID := utils.GetChatIdAndMsgIdAndUserID(t.update)
	logger.Info("command info", "userID", userID, "cmd", cmd)
	
	// check if at bot
	if (utils.GetChatType(t.update) == "group" || utils.GetChatType(t.update) == "supergroup") && *conf.BaseConfInfo.NeedATBOt {
		if !strings.Contains(t.update.Message.Text, "@"+t.bot.Self.UserName) {
			logger.Warn("not at bot", "userID", userID, "cmd", cmd)
			return
		}
	}
	
	switch cmd {
	case "chat":
		t.sendChatMessage()
	case "mode":
		t.sendModeConfigurationOptions()
	case "balance":
		t.showBalanceInfo()
	case "state":
		t.showStateInfo()
	case "clear":
		t.clearAllRecord()
	case "retry":
		t.retryLastQuestion()
	case "photo":
		t.sendImg()
	case "video":
		t.sendVideo()
	case "help":
		t.sendHelpConfigurationOptions()
	case "task":
		t.sendMultiAgent("task_empty_content")
	case "mcp":
		t.sendMultiAgent("mcp_empty_content")
	}
	
	if t.checkAdminUser() {
		switch cmd {
		case "addtoken":
			t.addToken()
		}
	}
}

// sendChatMessage response chat command to telegram
func (t *TelegramRobot) sendChatMessage() {
	chatId, msgID, _ := utils.GetChatIdAndMsgIdAndUserID(t.update)
	
	messageText := ""
	if t.update.Message != nil {
		messageText = t.update.Message.Text
		if messageText == "" && t.update.Message.Voice != nil && *conf.AudioConfInfo.AudioAppID != "" {
			audioContent := utils.GetAudioContent(t.update, t.bot)
			if audioContent == nil {
				logger.Warn("audio url empty")
				return
			}
			messageText = utils.FileRecognize(audioContent)
		}
		
		if messageText == "" && t.update.Message.Photo != nil {
			photoContent, err := utils.GetImageContent(utils.GetPhotoContent(t.update, t.bot))
			if err != nil {
				logger.Warn("get photo content err", "err", err)
				return
			}
			messageText = photoContent
		}
		
	} else {
		t.update.Message = new(tgbotapi.Message)
	}
	
	// Remove /chat and /chat@botUserName from the message
	content := utils.ReplaceCommand(messageText, "/chat", t.bot.Self.UserName)
	t.update.Message.Text = content
	
	if len(content) == 0 {
		err := utils.ForceReply(chatId, msgID, "chat_empty_content", t.bot)
		if err != nil {
			logger.Warn("force reply fail", "err", err)
		}
		return
	}
	
	// Reply to the chat content
	t.requestDeepseekAndResp(content)
}

// retryLastQuestion retry last question
func (t *TelegramRobot) retryLastQuestion() {
	chatId, msgId, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		t.requestDeepseekAndResp(records.AQs[len(records.AQs)-1].Question)
	} else {
		i18n.SendMsg(chatId, "last_question_fail", t.bot, nil, msgId)
	}
}

// clearAllRecord clear all record
func (t *TelegramRobot) clearAllRecord() {
	chatId, msgId, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	db.DeleteMsgRecord(userId)
	i18n.SendMsg(chatId, "delete_succ", t.bot, nil, msgId)
}

// addToken clear all record
func (t *TelegramRobot) addToken() {
	chatId, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(t.update)
	msg := utils.GetMessage(t.update)
	
	content := utils.ReplaceCommand(msg.Text, "/addtoken", t.bot.Self.UserName)
	splitContent := strings.Split(content, " ")
	
	db.AddAvailToken(int64(utils.ParseInt(splitContent[0])), utils.ParseInt(splitContent[1]))
	i18n.SendMsg(chatId, "add_token_succ", t.bot, nil, msgId)
}

// showBalanceInfo show balance info
func (t *TelegramRobot) showBalanceInfo() {
	chatId, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(t.update)
	
	if *conf.BaseConfInfo.Type != param.DeepSeek {
		i18n.SendMsg(chatId, "not_deepseek", t.bot, nil, msgId)
		return
	}
	
	balance := llm.GetBalanceInfo()
	
	// handle balance info msg
	msgContent := fmt.Sprintf(i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_title", nil), balance.IsAvailable)
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "balance_content", nil)
	
	for _, bInfo := range balance.BalanceInfos {
		msgContent += fmt.Sprintf(template, bInfo.Currency, bInfo.TotalBalance,
			bInfo.ToppedUpBalance, bInfo.GrantedBalance)
	}
	
	utils.SendMsg(chatId, msgContent, t.bot, msgId, tgbotapi.ModeMarkdown)
}

// showStateInfo show user's usage of token
func (t *TelegramRobot) showStateInfo() {
	chatId, msgId, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		return
	}
	
	if userInfo == nil {
		db.InsertUser(userId, godeepseek.DeepSeekChat)
		userInfo, err = db.GetUserByID(userId)
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
	
	template := i18n.GetMessage(*conf.BaseConfInfo.Lang, "state_content", nil)
	msgContent := fmt.Sprintf(template, userInfo.Token, todayTokey, weekToken, monthToken)
	utils.SendMsg(chatId, msgContent, t.bot, msgId, tgbotapi.ModeMarkdown)
}

// sendModeConfigurationOptions send config view
func (t *TelegramRobot) sendModeConfigurationOptions() {
	chatID, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(t.update)
	
	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
	switch *conf.BaseConfInfo.Type {
	case param.DeepSeek:
		if *conf.BaseConfInfo.CustomUrl == "" || *conf.BaseConfInfo.CustomUrl == "https://api.deepseek.com/" {
			for k := range param.DeepseekModels {
				inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(k, k),
				))
			}
		} else {
			inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(godeepseek.AzureDeepSeekR1, godeepseek.AzureDeepSeekR1),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(godeepseek.OpenRouterDeepSeekR1, godeepseek.OpenRouterDeepSeekR1),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(godeepseek.OpenRouterDeepSeekR1DistillLlama70B, godeepseek.OpenRouterDeepSeekR1DistillLlama70B),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(godeepseek.OpenRouterDeepSeekR1DistillLlama8B, godeepseek.OpenRouterDeepSeekR1DistillLlama8B),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(godeepseek.OpenRouterDeepSeekR1DistillQwen14B, godeepseek.OpenRouterDeepSeekR1DistillQwen14B),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(godeepseek.OpenRouterDeepSeekR1DistillQwen1_5B, godeepseek.OpenRouterDeepSeekR1DistillQwen1_5B),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(godeepseek.OpenRouterDeepSeekR1DistillQwen32B, godeepseek.OpenRouterDeepSeekR1DistillQwen32B),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("llama2", param.LLAVA),
				),
			)
		}
	case param.Gemini:
		for k := range param.GeminiModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.OpenAi:
		for k := range param.OpenAIModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.LLAVA:
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("llama2", param.LLAVA),
		))
	case param.OpenRouter:
		for k := range param.OpenRouterModelTypes {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Vol:
		// create inline button
		for k := range param.VolModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
		
	}
	
	inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
	
	i18n.SendMsg(chatID, "chat_mode", t.bot, &inlineKeyboard, msgId)
}

// sendHelpConfigurationOptions
func (t *TelegramRobot) sendHelpConfigurationOptions() {
	chatID, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(t.update)
	
	// create inline button
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("mode", "mode"),
			tgbotapi.NewInlineKeyboardButtonData("clear", "clear"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("balance", "balance"),
			tgbotapi.NewInlineKeyboardButtonData("state", "state"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("retry", "retry"),
			tgbotapi.NewInlineKeyboardButtonData("chat", "chat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("photo", "photo"),
			tgbotapi.NewInlineKeyboardButtonData("video", "video"),
		),
	)
	
	i18n.SendMsg(chatID, "command_notice", t.bot, &inlineKeyboard, msgId)
}

// handleCallbackQuery handle callback response
func (t *TelegramRobot) handleCallbackQuery() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleCommand panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	switch t.update.CallbackQuery.Data {
	case "mode":
		t.sendModeConfigurationOptions()
	case "balance":
		t.showBalanceInfo()
	case "clear":
		t.clearAllRecord()
	case "retry":
		t.retryLastQuestion()
	case "state":
		t.showStateInfo()
	case "photo":
		if t.update.CallbackQuery.Message.ReplyToMessage != nil {
			t.update.CallbackQuery.Message.MessageID = t.update.CallbackQuery.Message.ReplyToMessage.MessageID
		}
		t.sendImg()
	case "video":
		if t.update.CallbackQuery.Message.ReplyToMessage != nil {
			t.update.CallbackQuery.Message.MessageID = t.update.CallbackQuery.Message.ReplyToMessage.MessageID
		}
		t.sendVideo()
	case "chat":
		if t.update.CallbackQuery.Message.ReplyToMessage != nil {
			t.update.CallbackQuery.Message.MessageID = t.update.CallbackQuery.Message.ReplyToMessage.MessageID
		}
		t.sendChatMessage()
	default:
		if param.GeminiModels[t.update.CallbackQuery.Data] || param.OpenAIModels[t.update.CallbackQuery.Data] ||
			param.DeepseekModels[t.update.CallbackQuery.Data] || param.DeepseekLocalModels[t.update.CallbackQuery.Data] ||
			param.OpenRouterModels[t.update.CallbackQuery.Data] || param.VolModels[t.update.CallbackQuery.Data] {
			t.handleModeUpdate()
		}
		if param.OpenRouterModelTypes[t.update.CallbackQuery.Data] {
			chatID, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(t.update)
			inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
			for k := range param.OpenRouterModels {
				if strings.Contains(k, t.update.CallbackQuery.Data+"/") {
					inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(k, k),
					))
				}
			}
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
			i18n.SendMsg(chatID, "chat_mode", t.bot, &inlineKeyboard, msgId)
			
		}
	}
	
}

// handleModeUpdate handle mode update
func (t *TelegramRobot) handleModeUpdate() {
	userInfo, err := db.GetUserByID(t.update.CallbackQuery.From.ID)
	if err != nil {
		logger.Warn("get user fail", "userID", t.update.CallbackQuery.From.ID, "err", err)
		t.sendFailMessage()
		return
	}
	
	if userInfo != nil && userInfo.ID != 0 {
		err = db.UpdateUserMode(t.update.CallbackQuery.From.ID, t.update.CallbackQuery.Data)
		if err != nil {
			logger.Warn("update user fail", "userID", t.update.CallbackQuery.From.ID, "err", err)
			t.sendFailMessage()
			return
		}
	} else {
		_, err = db.InsertUser(t.update.CallbackQuery.From.ID, t.update.CallbackQuery.Data)
		if err != nil {
			logger.Warn("insert user fail", "userID", t.update.CallbackQuery.From.String(), "err", err)
			t.sendFailMessage()
			return
		}
	}
	
	// send response
	callback := tgbotapi.NewCallback(t.update.CallbackQuery.ID, t.update.CallbackQuery.Data)
	if _, err := t.bot.Request(callback); err != nil {
		logger.Warn("request callback fail", "err", err)
	}
	
	//utils.SendMsg(update.CallbackQuery.Message.Chat.ID,
	//	i18n.GetMessage(*conf.BaseConfInfo.Lang, "mode_choose", nil)+update.CallbackQuery.Data, bot, update.CallbackQuery.Message.MessageID)
}

// sendFailMessage send set mode fail msg
func (t *TelegramRobot) sendFailMessage() {
	callback := tgbotapi.NewCallback(t.update.CallbackQuery.ID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_mode", nil))
	if _, err := t.bot.Request(callback); err != nil {
		logger.Warn("request callback fail", "err", err)
	}
	
	i18n.SendMsg(t.update.CallbackQuery.Message.Chat.ID, "set_mode", t.bot, nil, t.update.CallbackQuery.Message.MessageID)
}

func (t *TelegramRobot) sendMultiAgent(agentType string) {
	if utils.CheckUserChatExceed(t.update, t.bot) {
		return
	}
	
	defer func() {
		utils.DecreaseUserChat(t.update)
	}()
	
	chatId, replyToMessageID, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	if t.checkUserTokenExceed() {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	prompt := ""
	if t.update.Message != nil {
		prompt = t.update.Message.Text
	}
	prompt = utils.ReplaceCommand(prompt, "/mcp", t.bot.Self.UserName)
	prompt = utils.ReplaceCommand(prompt, "/task", t.bot.Self.UserName)
	if len(prompt) == 0 {
		err := utils.ForceReply(chatId, replyToMessageID, agentType, t.bot)
		if err != nil {
			logger.Warn("force reply fail", "err", err)
		}
		return
	}
	
	// send response message
	messageChan := make(chan *param.MsgInfo)
	
	dpReq := &llm.DeepseekTaskReq{
		Content:     prompt,
		Update:      t.update,
		Bot:         t.bot,
		MessageChan: messageChan,
	}
	
	if agentType == "mcp_empty_content" {
		go dpReq.ExecuteMcp()
	} else {
		go dpReq.ExecuteTask()
	}
	
	go t.handleUpdate(messageChan)
}

// sendVideo send video to telegram
func (t *TelegramRobot) sendVideo() {
	if utils.CheckUserChatExceed(t.update, t.bot) {
		return
	}
	
	defer func() {
		utils.DecreaseUserChat(t.update)
	}()
	
	chatId, replyToMessageID, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	if t.checkUserTokenExceed() {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	prompt := ""
	if t.update.Message != nil {
		prompt = t.update.Message.Text
	}
	
	prompt = utils.ReplaceCommand(prompt, "/video", t.bot.Self.UserName)
	if len(prompt) == 0 {
		err := utils.ForceReply(chatId, replyToMessageID, "video_empty_content", t.bot)
		if err != nil {
			logger.Warn("force reply fail", "err", err)
		}
		return
	}
	
	thinkingMsgId := i18n.SendMsg(chatId, "thinking", t.bot, nil, replyToMessageID)
	
	var videoUrl string
	var videoContent []byte
	var err error
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		videoUrl, err = llm.GenerateVolVideo(prompt)
	case param.Gemini:
		videoContent, err = llm.GenerateGeminiVideo(prompt)
	default:
		err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
	}
	if err != nil {
		logger.Warn("generate video fail", "err", err)
		return
	}
	
	var video tgbotapi.InputMediaVideo
	if len(videoUrl) != 0 {
		video = tgbotapi.NewInputMediaVideo(tgbotapi.FileURL(videoUrl))
	} else if len(videoContent) != 0 {
		video = tgbotapi.NewInputMediaVideo(tgbotapi.FileBytes{
			Name:  "video.mp4",
			Bytes: videoContent,
		})
	}
	
	edit := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    chatId,
			MessageID: thinkingMsgId,
		},
		Media: video,
	}
	
	_, err = t.bot.Request(edit)
	if err != nil {
		logger.Warn("send video fail", "result", edit)
		return
	}
	
	db.InsertRecordInfo(&db.Record{
		UserId:    userId,
		Question:  prompt,
		Answer:    videoUrl,
		Token:     param.VideoTokenUsage,
		IsDeleted: 1,
	})
	return
}

// sendImg send img to telegram
func (t *TelegramRobot) sendImg() {
	if utils.CheckUserChatExceed(t.update, t.bot) {
		return
	}
	
	defer func() {
		utils.DecreaseUserChat(t.update)
	}()
	
	chatId, replyToMessageID, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	if t.checkUserTokenExceed() {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	prompt := ""
	if t.update.Message != nil {
		prompt = t.update.Message.Text
	}
	
	prompt = utils.ReplaceCommand(prompt, "/photo", t.bot.Self.UserName)
	if len(prompt) == 0 {
		err := utils.ForceReply(chatId, replyToMessageID, "photo_empty_content", t.bot)
		if err != nil {
			logger.Warn("force reply fail", "err", err)
		}
		return
	}
	
	thinkingMsgId := i18n.SendMsg(chatId, "thinking", t.bot, nil, replyToMessageID)
	
	var imageUrl string
	var imageContent []byte
	var err error
	
	switch *conf.BaseConfInfo.MediaType {
	case param.Vol:
		imageUrl, err = llm.GenerateVolImg(prompt)
	case param.OpenAi:
		imageUrl, err = llm.GenerateOpenAIImg(prompt)
	case param.Gemini:
		imageContent, err = llm.GenerateGeminiImg(prompt)
	default:
		err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
	}
	
	if err != nil {
		logger.Warn("generate image fail", "err", err)
		utils.SendMsg(chatId, err.Error(), t.bot, replyToMessageID, "")
		return
	}
	
	var photo tgbotapi.InputMediaPhoto
	if len(imageUrl) != 0 {
		photo = tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(imageUrl))
	} else if len(imageContent) != 0 {
		photo = tgbotapi.NewInputMediaPhoto(tgbotapi.FileBytes{
			Name:  "image.jpg",
			Bytes: imageContent,
		})
	}
	
	edit := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    chatId,
			MessageID: thinkingMsgId,
		},
		Media: photo,
	}
	
	_, err = t.bot.Request(edit)
	if err != nil {
		logger.Warn("send image fail", "result", edit)
		return
	}
	
	db.InsertRecordInfo(&db.Record{
		UserId:    userId,
		Question:  prompt,
		Answer:    imageUrl,
		Token:     param.ImageTokenUsage,
		IsDeleted: 1,
	})
	
	return
}

// checkUserAllow check use can use telegram bot or not
func (t *TelegramRobot) checkUserAllow() bool {
	if len(conf.BaseConfInfo.AllowedTelegramUserIds) == 0 {
		return true
	}
	if conf.BaseConfInfo.AllowedTelegramUserIds[0] {
		return false
	}
	
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	_, ok := conf.BaseConfInfo.AllowedTelegramUserIds[userId]
	return ok
}

func (t *TelegramRobot) checkGroupAllow() bool {
	chat := utils.GetChat(t.update)
	if chat == nil {
		return false
	}
	
	if chat.IsGroup() || chat.IsSuperGroup() { // 判断是否是群组或超级群组
		if len(conf.BaseConfInfo.AllowedTelegramGroupIds) == 0 {
			return true
		}
		if conf.BaseConfInfo.AllowedTelegramGroupIds[0] {
			return false
		}
		if _, ok := conf.BaseConfInfo.AllowedTelegramGroupIds[chat.ID]; ok {
			return true
		}
	}
	
	return false
}

// checkUserTokenExceed check use token exceeded
func (t *TelegramRobot) checkUserTokenExceed() bool {
	if *conf.BaseConfInfo.TokenPerUser == 0 {
		return false
	}
	
	chatId, msgId, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user info fail", "err", err)
		return false
	}
	
	if userInfo == nil {
		db.InsertUser(userId, godeepseek.DeepSeekChat)
		logger.Warn("get user info is nil")
		return false
	}
	
	if userInfo.Token >= userInfo.AvailToken {
		tpl := i18n.GetMessage(*conf.BaseConfInfo.Lang, "token_exceed", nil)
		content := fmt.Sprintf(tpl, userInfo.Token, userInfo.AvailToken-userInfo.Token, userInfo.AvailToken)
		utils.SendMsg(chatId, content, t.bot, msgId, tgbotapi.ModeMarkdown)
		return true
	}
	
	return false
}

// checkAdminUser check user is admin
func (t *TelegramRobot) checkAdminUser() bool {
	if len(conf.BaseConfInfo.AdminUserIds) == 0 {
		return false
	}
	
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(t.update)
	_, ok := conf.BaseConfInfo.AdminUserIds[userId]
	return ok
}

// ExecuteForceReply use force reply interact with user
func (t *TelegramRobot) ExecuteForceReply() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("ExecuteForceReply panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	switch t.update.Message.ReplyToMessage.Text {
	case i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_empty_content", nil):
		t.sendChatMessage()
	case i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_empty_content", nil):
		t.sendImg()
	case i18n.GetMessage(*conf.BaseConfInfo.Lang, "video_empty_content", nil):
		t.sendVideo()
	case i18n.GetMessage(*conf.BaseConfInfo.Lang, "task_empty_content", nil):
		t.sendMultiAgent("task_empty_content")
	case i18n.GetMessage(*conf.BaseConfInfo.Lang, "mcp_empty_content", nil):
		t.sendMultiAgent("mcp_empty_content")
	}
}
