package robot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Update tgbotapi.Update
	Bot    *tgbotapi.BotAPI
	
	Robot *Robot
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
			t.Robot = NewRobot(WithTelegramRobot(t))
			t.Robot.Exec()
		}
	}
}

func CreateBot() *tgbotapi.BotAPI {
	// 配置自定义 HTTP Client 并设置代理
	client := utils.GetTelegramProxyClient()
	
	var err error
	conf.BaseConfInfo.Bot, err = tgbotapi.NewBotAPIWithClient(*conf.BaseConfInfo.TelegramBotToken, tgbotapi.APIEndpoint, client)
	if err != nil {
		panic("Init bot fail" + err.Error())
	}
	
	if *logger.LogLevel == "debug" {
		conf.BaseConfInfo.Bot.Debug = true
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
	conf.BaseConfInfo.Bot.Send(cmdCfg)
	
	return conf.BaseConfInfo.Bot
}

func (t *TelegramRobot) Execute() {
	chatId, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	if t.handleCommandAndCallback() {
		return
	}
	// check whether you have new message
	if t.Update.Message != nil {
		if t.skipThisMsg() {
			logger.Warn("skip this msg", "msgId", msgId, "chat", chatId, "type", t.Update.Message.Chat.Type, "content", t.Update.Message.Text)
			return
		}
		t.requestDeepseekAndResp(t.Update.Message.Text)
	}
}

// requestDeepseekAndResp request deepseek api
func (t *TelegramRobot) requestDeepseekAndResp(content string) {
	chatId, replyToMessageID, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	if t.Robot.checkUserTokenExceed(chatId, replyToMessageID, userId) {
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
	chatId, msgId, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("GetContent panic err", "err", err, "stack", string(debug.Stack()))
			}
			utils.DecreaseUserChat(userId)
			close(messageChan)
		}()
		// check user chat exceed max count
		if utils.CheckUserChatExceed(userId) {
			t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
		
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
		utils.DecreaseUserChat(userId)
		close(messageChan)
	}()
	// check user chat exceed max count
	if utils.CheckUserChatExceed(userId) {
		t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
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
	
	chatId, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	parseMode := tgbotapi.ModeMarkdown
	
	tgMsgInfo := tgbotapi.NewMessage(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil))
	tgMsgInfo.ReplyToMessageID = msgId
	firstSendInfo, err := t.Bot.Send(tgMsgInfo)
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
			msg.MsgId = sendInfo.MessageID
		} else {
			updateMsg := tgbotapi.NewEditMessageText(chatId, msg.MsgId, msg.Content)
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
	chatType := t.GetChatType()
	if (chatType == "group" || chatType == "supergroup") && *conf.BaseConfInfo.NeedATBOt {
		if !strings.Contains(t.Update.Message.Text, "@"+t.Bot.Self.UserName) {
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
	
	if t.Robot.checkAdminUser(userID) {
		switch cmd {
		case "addtoken":
			t.addToken()
		}
	}
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
		err := utils.ForceReply(chatId, msgID, "chat_empty_content", t.Bot)
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
	chatId, msgId, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	records := db.GetMsgRecord(userId)
	if records != nil && len(records.AQs) > 0 {
		t.requestDeepseekAndResp(records.AQs[len(records.AQs)-1].Question)
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

// addToken clear all record
func (t *TelegramRobot) addToken() {
	chatId, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	msg := t.GetMessage()
	
	content := utils.ReplaceCommand(msg.Text, "/addtoken", t.Bot.Self.UserName)
	splitContent := strings.Split(content, " ")
	
	db.AddAvailToken(splitContent[0], utils.ParseInt(splitContent[1]))
	t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "add_token_succ", nil),
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
			t.handleModeUpdate()
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

// handleModeUpdate handle mode update
func (t *TelegramRobot) handleModeUpdate() {
	_, _, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Warn("get user fail", "userID", t.Update.CallbackQuery.From.ID, "err", err)
		t.sendFailMessage()
		return
	}
	
	if userInfo != nil && userInfo.ID != 0 {
		err = db.UpdateUserMode(t.Update.CallbackQuery.From.ID, t.Update.CallbackQuery.Data)
		if err != nil {
			logger.Warn("update user fail", "userID", t.Update.CallbackQuery.From.ID, "err", err)
			t.sendFailMessage()
			return
		}
	} else {
		_, err = db.InsertUser(userId, t.Update.CallbackQuery.Data)
		if err != nil {
			logger.Warn("insert user fail", "userID", t.Update.CallbackQuery.From.String(), "err", err)
			t.sendFailMessage()
			return
		}
	}
	
	// send response
	callback := tgbotapi.NewCallback(t.Update.CallbackQuery.ID, t.Update.CallbackQuery.Data)
	if _, err := t.Bot.Request(callback); err != nil {
		logger.Warn("request callback fail", "err", err)
	}
	
	//utils.SendMsg(update.CallbackQuery.Message.Chat.ID,
	//	i18n.GetMessage(*conf.BaseConfInfo.Lang, "mode_choose", nil)+update.CallbackQuery.Data, bot, update.CallbackQuery.Message.MessageID)
}

// sendFailMessage send set mode fail msg
func (t *TelegramRobot) sendFailMessage() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	callback := tgbotapi.NewCallback(t.Update.CallbackQuery.ID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_mode", nil))
	if _, err := t.Bot.Request(callback); err != nil {
		logger.Warn("request callback fail", "err", err)
	}
	t.Robot.SendMsg(chatID, i18n.GetMessage(*conf.BaseConfInfo.Lang, "set_mode", nil),
		msgId, tgbotapi.ModeMarkdown, nil)
}

func (t *TelegramRobot) sendMultiAgent(agentType string) {
	chatId, replyToMessageID, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	if utils.CheckUserChatExceed(userId) {
		t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	defer func() {
		utils.DecreaseUserChat(userId)
	}()
	
	if t.Robot.checkUserTokenExceed(chatId, replyToMessageID, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	prompt := ""
	if t.Update.Message != nil {
		prompt = t.Update.Message.Text
	}
	prompt = utils.ReplaceCommand(prompt, "/mcp", t.Bot.Self.UserName)
	prompt = utils.ReplaceCommand(prompt, "/task", t.Bot.Self.UserName)
	if len(prompt) == 0 {
		err := utils.ForceReply(chatId, replyToMessageID, agentType, t.Bot)
		if err != nil {
			logger.Warn("force reply fail", "err", err)
		}
		return
	}
	
	// send response message
	messageChan := make(chan *param.MsgInfo)
	
	dpReq := &llm.DeepseekTaskReq{
		Content:     prompt,
		UserId:      userId,
		ChatId:      chatId,
		MsgId:       replyToMessageID,
		MessageChan: messageChan,
	}
	
	if agentType == "mcp_empty_content" {
		go func() {
			err := dpReq.ExecuteMcp()
			if err != nil {
				t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			}
		}()
	} else {
		go func() {
			err := dpReq.ExecuteTask()
			if err != nil {
				t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			}
		}()
	}
	
	go t.handleUpdate(messageChan)
}

// sendVideo send video to telegram
func (t *TelegramRobot) sendVideo() {
	chatId, replyToMessageID, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	if utils.CheckUserChatExceed(userId) {
		t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	defer func() {
		utils.DecreaseUserChat(userId)
	}()
	
	if t.Robot.checkUserTokenExceed(chatId, replyToMessageID, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	prompt := ""
	if t.Update.Message != nil {
		prompt = t.Update.Message.Text
	}
	
	prompt = utils.ReplaceCommand(prompt, "/video", t.Bot.Self.UserName)
	if len(prompt) == 0 {
		err := utils.ForceReply(chatId, replyToMessageID, "video_empty_content", t.Bot)
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
	
	_, err = t.Bot.Request(edit)
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
	chatId, replyToMessageID, userId := t.Robot.GetChatIdAndMsgIdAndUserID()
	if utils.CheckUserChatExceed(userId) {
		t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "chat_exceed", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	defer func() {
		utils.DecreaseUserChat(userId)
	}()
	
	if t.Robot.checkUserTokenExceed(chatId, replyToMessageID, userId) {
		logger.Warn("user token exceed", "userID", userId)
		return
	}
	
	prompt := ""
	if t.Update.Message != nil {
		prompt = t.Update.Message.Text
	}
	
	prompt = utils.ReplaceCommand(prompt, "/photo", t.Bot.Self.UserName)
	if len(prompt) == 0 {
		err := utils.ForceReply(chatId, replyToMessageID, "photo_empty_content", t.Bot)
		if err != nil {
			logger.Warn("force reply fail", "err", err)
		}
		return
	}
	
	thinkingMsgId := t.Robot.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "thinking", nil),
		replyToMessageID, tgbotapi.ModeMarkdown, nil)
	
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
		t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
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
	
	_, err = t.Bot.Request(edit)
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
	if content == "" && t.Update.Message.Voice != nil && *conf.AudioConfInfo.AudioAppID != "" {
		audioContent := t.GetAudioContent()
		if audioContent == nil {
			logger.Warn("audio url empty")
			return "", errors.New("audio url empty")
		}
		content = utils.FileRecognize(audioContent)
	}
	
	if content == "" && t.Update.Message.Photo != nil {
		imageContent, err := utils.GetImageContent(t.GetPhotoContent())
		if err != nil {
			logger.Warn("get image content err", "err", err)
			return "", err
		}
		content = imageContent
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
	
	// 构造下载 URL
	downloadURL := file.Link(t.Bot.Token)
	
	transport := &http.Transport{}
	
	if *conf.BaseConfInfo.TelegramProxy != "" {
		proxy, err := url.Parse(*conf.BaseConfInfo.TelegramProxy)
		if err != nil {
			logger.Warn("parse proxy url fail", "err", err)
			return nil
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // 设置超时
	}
	
	// 通过代理下载
	resp, err := client.Get(downloadURL)
	if err != nil {
		logger.Warn("download fail", "err", err)
		return nil
	}
	defer resp.Body.Close()
	voice, err := io.ReadAll(resp.Body)
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
	
	client := utils.GetTelegramProxyClient()
	resp, err := client.Get(downloadURL)
	if err != nil {
		logger.Warn("download fail", "err", err)
		return nil
	}
	defer resp.Body.Close()
	photoContent, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("read response fail", "err", err)
		return nil
	}
	
	return photoContent
}

func (t *TelegramRobot) GetChat() *tgbotapi.Chat {
	if t.Update.Message != nil {
		return t.Update.Message.Chat
	}
	if t.Update.CallbackQuery != nil {
		return t.Update.CallbackQuery.Message.Chat
	}
	return nil
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

func (t *TelegramRobot) GetChatType() string {
	chat := t.GetChat()
	return chat.Type
}

func (t *TelegramRobot) CheckMsgIsCallback() bool {
	return t.Update.CallbackQuery != nil
}
