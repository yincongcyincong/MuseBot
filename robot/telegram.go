package robot

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/llm"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/rag"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/vectorstores"
)

type TelegramRobot struct {
	Update tgbotapi.Update
	Bot    *tgbotapi.BotAPI
	
	Robot *RobotInfo
}

func NewTelegramRobot(update tgbotapi.Update, bot *tgbotapi.BotAPI) *TelegramRobot {
	return &TelegramRobot{
		Update: update,
		Bot:    bot,
	}
}

// StartTelegramRobot start listen robot callback
func StartTelegramRobot() {
	for {
		bot := CreateBot()
		logger.Info("telegramBot Info", "username", bot.Self.UserName)
		
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		
		updates := bot.GetUpdatesChan(u)
		for update := range updates {
			t := NewTelegramRobot(update, bot)
			t.Robot = NewRobot(WithRobot(t))
			t.Robot.Exec()
		}
	}
}

func CreateBot() *tgbotapi.BotAPI {
	client := utils.GetRobotProxyClient()
	
	var err error
	var bot *tgbotapi.BotAPI
	bot, err = tgbotapi.NewBotAPIWithClient(*conf.BaseConfInfo.TelegramBotToken, tgbotapi.APIEndpoint, client)
	if err != nil {
		logger.Error("telegramBot Error", "error", err)
		return nil
	}
	
	if *logger.LogLevel == "debug" {
		bot.Debug = true
	}
	
	// set command
	cmdCfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "help",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.help.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "clear",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.clear.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "retry",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.retry.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "mode",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "balance",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.balance.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "state",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.state.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "photo",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.photo.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "video",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.video.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "chat",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.chat.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "task",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.task.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     "mcp",
			Description: i18n.GetMessage(*conf.BaseConfInfo.Lang, "commands.mcp.description", nil),
		},
	)
	bot.Send(cmdCfg)
	
	return bot
}

func (t *TelegramRobot) checkValid() bool {
	chatId, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	if t.handleCommandAndCallback() {
		return false
	}
	
	if t.Update.Message != nil {
		if t.skipThisMsg() {
			logger.Warn("skip this msg", "msgId", msgId, "chat", chatId, "type", t.Update.Message.Chat.Type, "content", t.Update.Message.Text)
			return false
		}
		
		return true
	}
	
	return false
}

func (t *TelegramRobot) getMsgContent() string {
	return t.Update.Message.Text
}

// requestLLMAndResp request deepseek api
func (t *TelegramRobot) requestLLMAndResp(content string) {
	t.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			t.executeChain(content)
		} else {
			t.executeLLM(content)
		}
	})
}

// executeChain use langchain to interact llm
func (t *TelegramRobot) executeChain(content string) {
	messageChan := make(chan *param.MsgInfo)
	chatId, msgId, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("GetContent panic err", "err", err, "stack", string(debug.Stack()))
			}
			close(messageChan)
		}()
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		
		text, err := t.GetContent(content)
		if err != nil {
			logger.Error("get content fail", "err", err)
			t.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
			return
		}
		
		dpLLM := rag.NewRag(llm.WithMessageChan(messageChan), llm.WithContent(content),
			llm.WithChatId(chatId), llm.WithMsgId(msgId),
			llm.WithUserId(userId))
		
		qaChain := chains.NewRetrievalQAFromLLM(
			dpLLM,
			vectorstores.ToRetriever(conf.RagConfInfo.Store, 3),
		)
		_, err = chains.Run(ctx, qaChain, text)
		if err != nil {
			logger.Warn("execute chain fail", "err", err)
			t.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		}
	}()
	
	// send response message
	go t.handleUpdate(messageChan)
	
}

// executeLLM directly interact llm
func (t *TelegramRobot) executeLLM(content string) {
	messageChan := make(chan *param.MsgInfo)
	// request DeepSeek API
	go t.callLLM(content, messageChan)
	
	// send response message
	go t.handleUpdate(messageChan)
	
}

func (t *TelegramRobot) callLLM(content string, messageChan chan *param.MsgInfo) {
	
	chatId, msgId, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	defer func() {
		if err := recover(); err != nil {
			logger.Error("GetContent panic err", "err", err, "stack", string(debug.Stack()))
		}
		close(messageChan)
	}()
	
	text, err := t.GetContent(content)
	if err != nil {
		logger.Error("get content fail", "err", err)
		t.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
		return
	}
	
	l := llm.NewLLM(llm.WithMessageChan(messageChan), llm.WithContent(text),
		llm.WithChatId(chatId), llm.WithMsgId(msgId),
		llm.WithUserId(userId),
		llm.WithTaskTools(&conf.AgentInfo{
			DeepseekTool:    conf.DeepseekTools,
			VolTool:         conf.VolTools,
			OpenAITools:     conf.OpenAITools,
			GeminiTools:     conf.GeminiTools,
			OpenRouterTools: conf.OpenRouterTools,
		}))
	
	err = l.CallLLM()
	if err != nil {
		logger.Error("get content fail", "err", err)
		t.Robot.SendMsg(chatId, err.Error(), msgId, "", nil)
	}
}

// handleUpdate handle robot msg sending
func (t *TelegramRobot) handleUpdate(messageChan chan *param.MsgInfo) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleUpdate panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	var msg *param.MsgInfo
	
	chatIdStr, msgIdStr, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	msgId := utils.ParseInt(msgIdStr)
	chatId := int64(utils.ParseInt(chatIdStr))
	parseMode := tgbotapi.ModeMarkdown
	
	tgMsgInfo := tgbotapi.NewMessage(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil))
	tgMsgInfo.ReplyToMessageID = msgId
	firstSendInfo, err := t.Bot.Send(tgMsgInfo)
	if err != nil {
		logger.Warn("Sending first message fail", "err", err)
	}
	
	for msg = range messageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from llm!"
		}
		if firstSendInfo.MessageID != 0 {
			msg.MsgId = strconv.Itoa(firstSendInfo.MessageID)
		}
		
		if msg.MsgId == "" && firstSendInfo.MessageID == 0 {
			tgMsgInfo = tgbotapi.NewMessage(chatId, msg.Content)
			tgMsgInfo.ReplyToMessageID = msgId
			tgMsgInfo.ParseMode = parseMode
			sendInfo, err := t.Bot.Send(tgMsgInfo)
			if err != nil {
				if sleepUtilNoLimit(msgId, err) {
					sendInfo, err = t.Bot.Send(tgMsgInfo)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					tgMsgInfo.ParseMode = ""
					sendInfo, err = t.Bot.Send(tgMsgInfo)
				} else {
					_, err = t.Bot.Send(tgMsgInfo)
				}
				if err != nil {
					logger.Warn("Error sending message:", "msgID", msgId, "err", err)
					continue
				}
			}
			msg.MsgId = strconv.Itoa(sendInfo.MessageID)
		} else {
			updateMsg := tgbotapi.NewEditMessageText(chatId, utils.ParseInt(msg.MsgId), msg.Content)
			updateMsg.ParseMode = parseMode
			_, err = t.Bot.Send(updateMsg)
			if err != nil {
				// try again
				if sleepUtilNoLimit(msgId, err) {
					_, err = t.Bot.Send(updateMsg)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					updateMsg.ParseMode = ""
					_, err = t.Bot.Send(updateMsg)
				} else {
					_, err = t.Bot.Send(updateMsg)
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
	if t.Update.Message != nil && t.Update.Message.IsCommand() {
		go t.handleCommand()
		return true
	}
	
	if t.Update.CallbackQuery != nil {
		go t.handleCallbackQuery()
		return true
	}
	
	if t.Update.Message != nil && t.Update.Message.ReplyToMessage != nil && t.Update.Message.ReplyToMessage.From != nil &&
		t.Update.Message.ReplyToMessage.From.UserName == t.Bot.Self.UserName {
		go t.ExecuteForceReply()
		return true
	}
	
	return false
}

// skipThisMsg check if msg trigger llm
func (t *TelegramRobot) skipThisMsg() bool {
	if t.Update.Message.Chat.Type == "private" {
		if strings.TrimSpace(t.Update.Message.Text) == "" &&
			t.Update.Message.Voice == nil && t.Update.Message.Photo == nil {
			return true
		}
		
		return false
	} else {
		if strings.TrimSpace(strings.ReplaceAll(t.Update.Message.Text, "@"+t.Bot.Self.UserName, "")) == "" &&
			t.Update.Message.Voice == nil {
			return true
		}
		
		if !strings.Contains(t.Update.Message.Text, "@"+t.Bot.Self.UserName) {
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
	
	cmd := t.Update.Message.Command()
	_, _, userID := t.Robot.GetChatIdAndMsgIdAndUserID()
	logger.Info("command info", "userID", userID, "cmd", cmd)
	
	// check if at bot
	chatType := t.GetMessage().Chat.Type
	if (chatType == "group" || chatType == "supergroup") && *conf.BaseConfInfo.NeedATBOt {
		if !strings.Contains(t.Update.Message.Text, "@"+t.Bot.Self.UserName) {
			logger.Warn("not at bot", "userID", userID, "cmd", cmd)
			return
		}
	}
	
	t.Robot.ExecCmd(cmd)
}

// sendChatMessage response chat command to telegram
func (t *TelegramRobot) sendChatMessage() {
	chatId, msgID, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	messageText := ""
	var err error
	if t.Update.Message != nil {
		messageText = t.Update.Message.Text
		messageText, err = t.GetContent(messageText)
		if err != nil {
			logger.Warn("GetContent error", "err", err)
			return
		}
	} else {
		t.Update.Message = new(tgbotapi.Message)
	}
	
	// Remove /chat and /chat@botUserName from the message
	content := utils.ReplaceCommand(messageText, "/chat", t.Bot.Self.UserName)
	t.Update.Message.Text = content
	
	if len(content) == 0 {
		err := utils.ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(msgID), "chat_empty_content", t.Bot)
		if err != nil {
			logger.Warn("force reply fail", "err", err)
		}
		return
	}
	
	// Reply to the chat content
	t.requestLLMAndResp(content)
}

// retryLastQuestion retry last question
func (t *TelegramRobot) retryLastQuestion() {
	chatId, msgId, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		t.requestLLMAndResp(records.AQs[len(records.AQs)-1].Question)
	} else {
		t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "last_question_fail", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
	}
}

// clearAllRecord clear all record
func (t *TelegramRobot) clearAllRecord() {
	chatId, msgId, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	db.DeleteMsgRecord(userId)
	t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "delete_succ", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
}

// showBalanceInfo show balance info
func (t *TelegramRobot) showBalanceInfo() {
	chatId, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	if *conf.BaseConfInfo.Type != param.DeepSeek {
		t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "not_deepseek", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
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
	
	t.Robot.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
}

// showStateInfo show user's usage of token
func (t *TelegramRobot) showStateInfo() {
	chatId, msgId, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		t.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
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
	t.Robot.SendMsg(chatId, msgContent, msgId, tgbotapi.ModeMarkdown, nil)
}

// sendModeConfigurationOptions send config view
func (t *TelegramRobot) sendModeConfigurationOptions() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	
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
	
	t.Robot.SendMsg(chatID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_mode", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

// sendHelpConfigurationOptions
func (t *TelegramRobot) sendHelpConfigurationOptions() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	
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
	
	t.Robot.SendMsg(chatID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "command_notice", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

// handleCallbackQuery handle callback response
func (t *TelegramRobot) handleCallbackQuery() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("handleCommand panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	switch t.Update.CallbackQuery.Data {
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
		if t.Update.CallbackQuery.Message.ReplyToMessage != nil {
			t.Update.CallbackQuery.Message.MessageID = t.Update.CallbackQuery.Message.ReplyToMessage.MessageID
		}
		t.sendImg()
	case "video":
		if t.Update.CallbackQuery.Message.ReplyToMessage != nil {
			t.Update.CallbackQuery.Message.MessageID = t.Update.CallbackQuery.Message.ReplyToMessage.MessageID
		}
		t.sendVideo()
	case "chat":
		if t.Update.CallbackQuery.Message.ReplyToMessage != nil {
			t.Update.CallbackQuery.Message.MessageID = t.Update.CallbackQuery.Message.ReplyToMessage.MessageID
		}
		t.sendChatMessage()
	default:
		if param.GeminiModels[t.Update.CallbackQuery.Data] || param.OpenAIModels[t.Update.CallbackQuery.Data] ||
			param.DeepseekModels[t.Update.CallbackQuery.Data] || param.DeepseekLocalModels[t.Update.CallbackQuery.Data] ||
			param.OpenRouterModels[t.Update.CallbackQuery.Data] || param.VolModels[t.Update.CallbackQuery.Data] {
			t.Robot.handleModeUpdate(t.Update.CallbackQuery.Data)
		}
		if param.OpenRouterModelTypes[t.Update.CallbackQuery.Data] {
			chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
			inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
			for k := range param.OpenRouterModels {
				if strings.Contains(k, t.Update.CallbackQuery.Data+"/") {
					inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(k, k),
					))
				}
			}
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
			t.Robot.SendMsg(chatID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_mode", nil),
				msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
			
		}
	}
	
}

func (t *TelegramRobot) sendMultiAgent(agentType string) {
	t.Robot.TalkingPreCheck(func() {
		chatId, replyToMessageID, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := ""
		if t.Update.Message != nil {
			prompt = t.Update.Message.Text
		}
		prompt = utils.ReplaceCommand(prompt, "/mcp", t.Bot.Self.UserName)
		prompt = utils.ReplaceCommand(prompt, "/task", t.Bot.Self.UserName)
		if len(prompt) == 0 {
			err := utils.ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(replyToMessageID), agentType, t.Bot)
			if err != nil {
				logger.Warn("force reply fail", "err", err)
			}
			return
		}
		
		// send response message
		messageChan := make(chan *param.MsgInfo)
		
		dpReq := &llm.LLMTaskReq{
			Content:     prompt,
			UserId:      userId,
			ChatId:      chatId,
			MsgId:       replyToMessageID,
			MessageChan: messageChan,
		}
		
		if agentType == "mcp_empty_content" {
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error("handleCommand panic err", "err", err, "stack", string(debug.Stack()))
					}
					close(messageChan)
				}()
				
				err := dpReq.ExecuteMcp()
				if err != nil {
					t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
				}
			}()
		} else {
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error("handleCommand panic err", "err", err, "stack", string(debug.Stack()))
					}
					close(messageChan)
				}()
				
				err := dpReq.ExecuteTask()
				if err != nil {
					t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
				}
			}()
		}
		
		go t.handleUpdate(messageChan)
	})
	
}

// sendVideo send video to telegram
func (t *TelegramRobot) sendVideo() {
	
	t.Robot.TalkingPreCheck(func() {
		chatId, replyToMessageID, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := ""
		if t.Update.Message != nil {
			prompt = t.Update.Message.Text
		}
		
		prompt = utils.ReplaceCommand(prompt, "/video", t.Bot.Self.UserName)
		if len(prompt) == 0 {
			err := utils.ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(replyToMessageID), "video_empty_content", t.Bot)
			if err != nil {
				logger.Warn("force reply fail", "err", err)
			}
			return
		}
		
		thinkingMsgId := t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		
		var videoUrl string
		var videoContent []byte
		var err error
		var totalToken int
		mode := *conf.BaseConfInfo.MediaType
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			videoUrl, totalToken, err = llm.GenerateVolVideo(prompt)
		case param.Gemini:
			videoContent, totalToken, err = llm.GenerateGeminiVideo(prompt)
		default:
			err = fmt.Errorf("unsupported type: %s", *conf.BaseConfInfo.MediaType)
		}
		if err != nil {
			logger.Warn("generate video fail", "err", err)
			t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
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
				ChatID:    int64(utils.ParseInt(chatId)),
				MessageID: utils.ParseInt(thinkingMsgId),
			},
			Media: video,
		}
		
		_, err = t.Bot.Request(edit)
		if err != nil {
			logger.Warn("send video fail", "result", edit)
			t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			return
		}
		
		if len(videoContent) == 0 {
			videoContent, err = utils.DownloadFile(videoUrl)
			if err != nil {
				logger.Warn("download video fail", "err", err)
				return
			}
		}
		
		base64Content := base64.StdEncoding.EncodeToString(videoContent)
		dataURI := fmt.Sprintf("data:video/%s;base64,%s", utils.DetectVideoMimeType(videoContent), base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.VideoRecordType,
			Mode:       mode,
		})
	})
}

// sendImg send img to telegram
func (t *TelegramRobot) sendImg() {
	t.Robot.TalkingPreCheck(func() {
		chatId, replyToMessageID, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := ""
		if t.Update.Message != nil {
			prompt = t.Update.Message.Text
		}
		prompt = utils.ReplaceCommand(prompt, "/photo", t.Bot.Self.UserName)
		if prompt == "" && t.Update.Message != nil && len(t.Update.Message.Photo) > 0 {
			prompt = t.Update.Message.Caption
		}
		
		if len(prompt) == 0 {
			err := utils.ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(replyToMessageID), "photo_empty_content", t.Bot)
			if err != nil {
				logger.Warn("force reply fail", "err", err)
			}
			return
		}
		
		var err error
		lastImageContent := t.GetPhotoContent()
		if len(lastImageContent) == 0 {
			lastImageContent, err = t.Robot.GetLastImageContent()
			if err != nil {
				logger.Warn("get last image record fail", "err", err)
			}
		}
		
		thinkingMsgId := t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		
		var imageUrl string
		var imageContent []byte
		var totalToken int
		mode := *conf.BaseConfInfo.MediaType
		switch *conf.BaseConfInfo.MediaType {
		case param.Vol:
			imageUrl, totalToken, err = llm.GenerateVolImg(prompt, lastImageContent)
		case param.OpenAi:
			imageContent, totalToken, err = llm.GenerateOpenAIImg(prompt, lastImageContent)
		case param.Gemini:
			imageContent, totalToken, err = llm.GenerateGeminiImg(prompt, lastImageContent)
		default:
			err = fmt.Errorf("unsupported media type: %s", *conf.BaseConfInfo.MediaType)
		}
		
		if err != nil {
			logger.Warn("generate image fail", "err", err)
			t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			return
		}
		
		var photo tgbotapi.InputMediaPhoto
		if len(imageContent) == 0 {
			imageContent, err = utils.DownloadFile(imageUrl)
			if err != nil {
				logger.Warn("download image fail", "err", err)
				return
			}
		}
		
		img, _, err := image.Decode(bytes.NewReader(imageContent))
		if err != nil {
			logger.Error("decode image fail", "err", err)
			return
		}
		resizedImg := imaging.Fit(img, 1280, 1280, imaging.Lanczos)
		var buf bytes.Buffer
		
		err = imaging.Encode(&buf, resizedImg, imaging.JPEG)
		if err != nil {
			logger.Error("encode image fail", "err", err)
			return
		}
		
		imageContent = buf.Bytes()
		
		photo = tgbotapi.NewInputMediaPhoto(tgbotapi.FileBytes{
			Name:  "image." + utils.DetectImageFormat(imageContent),
			Bytes: imageContent,
		})
		
		edit := tgbotapi.EditMessageMediaConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:    int64(utils.ParseInt(chatId)),
				MessageID: utils.ParseInt(thinkingMsgId),
			},
			Media: photo,
		}
		
		_, err = t.Bot.Request(edit)
		if err != nil {
			logger.Warn("send image fail", "result", edit)
			t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			return
		}
		
		base64Content := base64.StdEncoding.EncodeToString(imageContent)
		dataURI := fmt.Sprintf("data:image/%s;base64,%s", utils.DetectImageFormat(imageContent), base64Content)
		
		db.InsertRecordInfo(&db.Record{
			UserId:     userId,
			Question:   prompt,
			Answer:     dataURI,
			Token:      totalToken,
			IsDeleted:  0,
			RecordType: param.ImageRecordType,
			Mode:       mode,
		})
	})
}

// ExecuteForceReply use force reply interact with user
func (t *TelegramRobot) ExecuteForceReply() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("ExecuteForceReply panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	switch t.Update.Message.ReplyToMessage.Text {
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

func (t *TelegramRobot) GetContent(content string) (string, error) {
	var err error
	if content == "" && t.Update.Message.Voice != nil && *conf.AudioConfInfo.AudioAppID != "" {
		audioContent := t.GetAudioContent()
		if audioContent == nil {
			logger.Warn("audio url empty")
			return "", errors.New("audio url empty")
		}
		
		content, err = t.Robot.GetAudioContent(audioContent)
		if err != nil {
			logger.Warn("generate text fail", "err", err)
			return "", err
		}
		
	}
	
	if content == "" && t.Update.Message.Photo != nil {
		content, err = t.Robot.GetImageContent(t.GetPhotoContent())
		if err != nil {
			logger.Warn("get image content err", "err", err)
			return "", err
		}
	}
	
	if content == "" {
		logger.Warn("content empty")
		return "", errors.New("content empty")
	}
	
	text := strings.ReplaceAll(content, "@"+t.Bot.Self.UserName, "")
	return text, nil
}

func (t *TelegramRobot) GetAudioContent() []byte {
	if t.Update.Message == nil || t.Update.Message.Voice == nil {
		return nil
	}
	
	fileID := t.Update.Message.Voice.FileID
	file, err := t.Bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		logger.Warn("get file fail", "err", err)
		return nil
	}
	
	downloadURL := file.Link(t.Bot.Token)
	voice, err := utils.DownloadFile(downloadURL)
	if err != nil {
		logger.Warn("read response fail", "err", err)
		return nil
	}
	return voice
}

func (t *TelegramRobot) GetPhotoContent() []byte {
	if t.Update.Message == nil || t.Update.Message.Photo == nil {
		return nil
	}
	
	var photo tgbotapi.PhotoSize
	for i := len(t.Update.Message.Photo) - 1; i >= 0; i-- {
		if t.Update.Message.Photo[i].FileSize < 8*1024*1024 {
			photo = t.Update.Message.Photo[i]
			break
		}
	}
	
	fileID := photo.FileID
	file, err := t.Bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		logger.Warn("get file fail", "err", err)
		return nil
	}
	
	downloadURL := file.Link(t.Bot.Token)
	photoContent, err := utils.DownloadFile(downloadURL)
	if err != nil {
		logger.Warn("read response fail", "err", err)
		return nil
	}
	
	return photoContent
}

func (t *TelegramRobot) GetMessage() *tgbotapi.Message {
	if t.Update.Message != nil {
		return t.Update.Message
	}
	if t.Update.CallbackQuery != nil {
		return t.Update.CallbackQuery.Message
	}
	return nil
}
