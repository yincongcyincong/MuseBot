package robot

import (
	"bytes"
	"context"
	"errors"
	"image"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	
	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var TelegramBot *tgbotapi.BotAPI

type TelegramRobot struct {
	Update tgbotapi.Update
	Bot    *tgbotapi.BotAPI
	
	Robot        *RobotInfo
	Prompt       string
	Command      string
	ImageContent []byte
	AudioContent []byte
	UserName     string
}

func NewTelegramRobot(update tgbotapi.Update, bot *tgbotapi.BotAPI) *TelegramRobot {
	metrics.AppRequestCount.WithLabelValues("telegram").Inc()
	t := &TelegramRobot{
		Update: update,
		Bot:    bot,
	}
	
	if update.Message != nil && update.Message.From != nil {
		t.UserName = update.Message.From.UserName
	}
	if update.CallbackQuery != nil && update.CallbackQuery.From != nil {
		t.UserName = update.CallbackQuery.From.UserName
	}
	
	return t
}

// StartTelegramRobot start listen robot callback
func StartTelegramRobot(ctx context.Context) {
	
	defer func() {
		if TelegramBot != nil {
			TelegramBot.StopReceivingUpdates()
		}
		if err := recover(); err != nil {
			logger.ErrorCtx(ctx, "StartTelegramRobot panic", "err", err, "stack", string(debug.Stack()))
			StartTelegramRobot(ctx)
		}
	}()
	
	for {
		TelegramBot = CreateBot(ctx)
		logger.InfoCtx(ctx, "telegramBot Info", "username", TelegramBot.Self.UserName)
		
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		
		updates := TelegramBot.GetUpdatesChan(u)
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-updates:
				t := NewTelegramRobot(update, TelegramBot)
				t.Robot = NewRobot(WithRobot(t))
				t.Robot.Exec()
			}
		}
	}
}

func CreateBot(ctx context.Context) *tgbotapi.BotAPI {
	client := utils.GetRobotProxyClient()
	
	var err error
	var bot *tgbotapi.BotAPI
	bot, err = tgbotapi.NewBotAPIWithClient(*conf.BaseConfInfo.TelegramBotToken, tgbotapi.APIEndpoint, client)
	if err != nil {
		logger.ErrorCtx(ctx, "telegramBot Error", "error", err)
		return nil
	}
	
	if *logger.LogLevel == "debug" {
		bot.Debug = true
	}
	
	// set command
	cmdCfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     param.Help,
			Description: i18n.GetMessage("commands.help.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Clear,
			Description: i18n.GetMessage("commands.clear.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Retry,
			Description: i18n.GetMessage("commands.retry.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.TxtModel,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.PhotoModel,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.RecModel,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.TtsModel,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Mode,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.VideoModel,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.PhotoType,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.VideoType,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.RecType,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.TtsType,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.TxtType,
			Description: i18n.GetMessage("commands.mode.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.State,
			Description: i18n.GetMessage("commands.state.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Photo,
			Description: i18n.GetMessage("commands.photo.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.EditPhoto,
			Description: i18n.GetMessage("commands.edit_photo.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Video,
			Description: i18n.GetMessage("commands.video.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Chat,
			Description: i18n.GetMessage("commands.chat.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Task,
			Description: i18n.GetMessage("commands.task.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.Mcp,
			Description: i18n.GetMessage("commands.mcp.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.CronList,
			Description: i18n.GetMessage("commands.cron.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.CronDel,
			Description: i18n.GetMessage("commands.cron.description", nil),
		},
		tgbotapi.BotCommand{
			Command:     param.CronDel,
			Description: i18n.GetMessage("commands.cron.description", nil),
		},
	)
	bot.Send(cmdCfg)
	
	return bot
}

func (t *TelegramRobot) checkValid() bool {
	chatId, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	if t.Update.CallbackQuery != nil {
		return true
	}
	
	if t.Update.Message != nil {
		if t.skipThisMsg() {
			logger.WarnCtx(t.Robot.Ctx, "skip this msg", "msgId", msgId, "chat", chatId,
				"type", t.getMessage().Chat.Type, "content", t.getMsgContent())
			return false
		}
		
		t.Command, t.Prompt = ParseCommand(t.getMsgContent())
		if t.Update.Message.IsCommand() {
			t.Command = t.Update.Message.Command()
		}
		
		t.ImageContent = t.GetPhotoContent()
		t.AudioContent = t.GetAudioContent()
		var err error
		if t.AudioContent != nil {
			t.Prompt, err = t.Robot.GetAudioContent(t.AudioContent)
			if err != nil {
				logger.WarnCtx(t.Robot.Ctx, "convert audio to text fail", "err", err)
				t.Robot.SendMsg(chatId, err.Error(), msgId, tgbotapi.ModeMarkdown, nil)
				return false
			}
		}
	}
	
	return true
}

func (t *TelegramRobot) getMsgContent() string {
	content := ""
	if t.getMessage() != nil {
		content = t.getMessage().Text
	}
	
	if t.Update.CallbackQuery != nil {
		return ""
	}
	
	if content == "" && t.getMessage() != nil {
		content = t.getMessage().Caption
	}
	return content
}

// requestLLMAndResp request deepseek api
func (t *TelegramRobot) requestLLM(content string) {
	if t.handleCommandAndCallback() {
		return
	}
	
	t.sendChatMessage()
}

// executeChain use langchain to interact llm
func (t *TelegramRobot) executeChain() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	
	go t.Robot.ExecChain(t.Prompt, messageChan)
	
	// send response message
	go t.Robot.HandleUpdate(messageChan, "ogg_opus")
	
}

// executeLLM directly interact llm
func (t *TelegramRobot) executeLLM() {
	messageChan := &MsgChan{
		NormalMessageChan: make(chan *param.MsgInfo),
	}
	go t.Robot.ExecLLM(t.Prompt, messageChan)
	
	// send response message
	go t.Robot.HandleUpdate(messageChan, "ogg_opus")
	
}

// sleepUtilNoLimit handle "Too Many Requests" error
func (t *TelegramRobot) sleepUtilNoLimit(msgId int, err error) bool {
	var apiErr *tgbotapi.Error
	if errors.As(err, &apiErr) && apiErr.Message == "Too Many Requests" {
		waitTime := time.Duration(apiErr.RetryAfter) * time.Second
		logger.WarnCtx(t.Robot.Ctx, "Rate limited. Retrying after", "msgID", msgId, "waitTime", waitTime)
		time.Sleep(waitTime)
		return true
	}
	
	return false
}

// handleCommandAndCallback telegram command and callback function
func (t *TelegramRobot) handleCommandAndCallback() bool {
	// if it's command, directly
	if t.Update.Message != nil && (t.Command != "" || t.Update.Message.IsCommand()) {
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
	if t.Robot.cs.SkipCheck {
		return false
	}
	
	if t.Update.Message.ReplyToMessage != nil && t.Update.Message.ReplyToMessage.From != nil &&
		t.Update.Message.ReplyToMessage.From.UserName == t.Bot.Self.UserName {
		return false
	}
	
	if t.Update.Message.Chat.Type == "private" {
		if strings.TrimSpace(t.getMsgContent()) == "" &&
			t.Update.Message.Voice == nil && t.Update.Message.Photo == nil {
			return true
		}
		
		return false
	} else {
		if strings.TrimSpace(strings.ReplaceAll(t.getMsgContent(), "@"+t.Bot.Self.UserName, "")) == "" &&
			t.Update.Message.Voice == nil {
			return true
		}
		
		if !strings.Contains(t.getMsgContent(), "@"+t.Bot.Self.UserName) {
			return true
		}
	}
	
	return false
}

// handleCommand handle multiple commands
func (t *TelegramRobot) handleCommand() {
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorCtx(t.Robot.Ctx, "handleCommand panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	if t.Command == "" {
		t.Command = t.Update.Message.Command()
	}
	
	t.Robot.ExecCmd(t.Command, t.sendChatMessage, t.changeModel, t.changeType)
}

// sendChatMessage response chat command to telegram
func (t *TelegramRobot) sendChatMessage() {
	chatId, msgID, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	messageText := t.Prompt
	if t.Update.Message != nil {
		if messageText == "" {
			messageText = t.Update.Message.Caption
		}
	} else {
		t.Update.Message = new(tgbotapi.Message)
	}
	
	// Remove /chat and /chat@botUserName from the message
	content := utils.ReplaceCommand(messageText, "/chat", t.Bot.Self.UserName)
	t.Update.Message.Text = content
	t.Prompt = content
	
	if len(content) == 0 {
		if t.ImageContent != nil {
			t.Robot.saveRecord(t.ImageContent, t.ImageContent, param.ImageRecordType, 0)
		}
		err := ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(msgID), "chat_empty_content", t.Bot)
		if err != nil {
			logger.WarnCtx(t.Robot.Ctx, "force reply fail", "err", err)
		}
		return
	}
	
	// Reply to the chat content
	t.Robot.TalkingPreCheck(func() {
		if conf.RagConfInfo.Store != nil {
			t.executeChain()
		} else {
			t.executeLLM()
		}
	})
}

func (t *TelegramRobot) changeType(ty string) {
	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
	switch ty {
	case param.TxtType, "/" + param.TxtType:
		for _, k := range utils.GetAvailTxtType() {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	
	case param.PhotoType, "/" + param.PhotoType:
		for _, k := range utils.GetAvailImgType() {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	
	case param.VideoType, "/" + param.VideoType:
		for _, k := range utils.GetAvailVideoType() {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.RecType, "/" + param.RecType:
		for _, k := range utils.GetAvailRecType() {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.TtsType, "/" + param.TtsType:
		for _, k := range utils.GetAvailTTSType() {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	}
	
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
	t.Robot.SendMsg(chatID, i18n.GetMessage("chat_mode", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

// sendModeConfigurationOptions send config view
func (t *TelegramRobot) changeModel(ty string) {
	
	switch ty {
	case param.TxtModel, "/" + param.TxtModel:
		if t.getPrompt() != "" {
			t.Robot.handleModelUpdate(&RobotModel{TxtModel: t.Prompt})
			return
		}
		t.showTxtModel(ty)
	
	case param.PhotoModel, "/" + param.PhotoModel:
		if t.getPrompt() != "" {
			t.Robot.handleModelUpdate(&RobotModel{ImgModel: t.Prompt})
			return
		}
		t.showImageModel()
	
	case param.VideoModel, "/" + param.VideoModel:
		if t.getPrompt() != "" {
			t.Robot.handleModelUpdate(&RobotModel{VideoModel: t.Prompt})
			return
		}
		t.showVideoModel()
	case param.RecModel, "/" + param.RecModel:
		if t.getPrompt() != "" {
			t.Robot.handleModelUpdate(&RobotModel{RecModel: t.Prompt})
			return
		}
		t.showRecModel()
	case param.TtsModel, "/" + param.TtsModel:
		if t.getPrompt() != "" {
			t.Robot.handleModelUpdate(&RobotModel{TTSModel: t.Prompt})
			return
		}
		t.showTTSModel()
	}
}

func (t *TelegramRobot) showTxtModel(ty string) {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
	switch utils.GetTxtType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
	case param.DeepSeek:
		for k := range param.DeepseekModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Gemini:
		for k := range param.GeminiModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Aliyun:
		for k := range param.AliyunModel {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.OpenRouter, param.AI302, param.Ollama, param.OpenAi:
		if t.Prompt != "" {
			t.Robot.handleModelUpdate(&RobotModel{TxtType: t.Prompt})
			return
		}
		switch utils.GetTxtType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
		case param.OpenAi:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link":    "https://platform.openai.com/",
				"command": ty,
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.AI302:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link":    "https://302.ai/",
				"command": ty,
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenRouter:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link":    "https://openrouter.ai/",
				"command": ty,
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.Ollama:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link":    "https://ollama.com/",
				"command": ty,
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		}
		
		return
	case param.Vol:
		// create inline button
		for k := range param.VolModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
		
	}
	
	inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
	
	t.Robot.SendMsg(chatID, i18n.GetMessage("chat_mode", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

func (t *TelegramRobot) showImageModel() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
	
	switch utils.GetImgType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiImageModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Aliyun:
		for k := range param.AliyunImageModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.OpenRouter, param.AI302, param.Ollama, param.OpenAi:
		switch utils.GetImgType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
		case param.AI302:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenRouter:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://openrouter.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.Ollama:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://ollama.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		case param.OpenAi:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://platform.openai.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
		}
		
		return
	case param.Vol:
		for k := range param.VolImageModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	}
	
	inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
	
	t.Robot.SendMsg(chatID, i18n.GetMessage("chat_mode", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

func (t *TelegramRobot) showVideoModel() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
	
	switch utils.GetVideoType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiVideoModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Aliyun:
		for k := range param.AliyunVideoModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Vol:
		for k := range param.VolVideoModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.AI302:
		switch utils.GetVideoType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
		case param.AI302:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}
	
	inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
	
	t.Robot.SendMsg(chatID, i18n.GetMessage("chat_mode", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

func (t *TelegramRobot) showRecModel() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
	
	switch utils.GetRecType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiRecModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Aliyun:
		for k := range param.AliyunRecModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Vol:
		for k := range param.VolRecModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.AI302, param.OpenAi:
		switch utils.GetRecType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
		case param.AI302:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://302.ai/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		case param.OpenAi:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://platform.openai.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}
	
	inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
	
	t.Robot.SendMsg(chatID, i18n.GetMessage("chat_mode", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

func (t *TelegramRobot) showTTSModel() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	inlineButton := make([][]tgbotapi.InlineKeyboardButton, 0)
	
	switch utils.GetTTSType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
	case param.Gemini:
		for k := range param.GeminiTTSModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Aliyun:
		for k := range param.AliyunTTSModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.Vol:
		for k := range param.VolTTSModels {
			inlineButton = append(inlineButton, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(k, k),
			))
		}
	case param.OpenAi:
		switch utils.GetTTSType(db.GetCtxUserInfo(t.Robot.Ctx).LLMConfigRaw) {
		case param.OpenAi:
			t.Robot.SendMsg(chatID, i18n.GetMessage("mix_mode_choose", map[string]interface{}{
				"link": "https://platform.openai.com/",
			}),
				msgId, tgbotapi.ModeMarkdown, nil)
			return
		}
	}
	
	inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(inlineButton...)
	
	t.Robot.SendMsg(chatID, i18n.GetMessage("chat_mode", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

// sendHelpConfigurationOptions
func (t *TelegramRobot) sendHelpConfigurationOptions() {
	chatID, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	
	// create inline button
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("clear", "clear"),
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
	
	t.Robot.SendMsg(chatID, i18n.GetMessage("command_notice", nil),
		msgId, tgbotapi.ModeMarkdown, &inlineKeyboard)
}

// handleCallbackQuery handle callback response
func (t *TelegramRobot) handleCallbackQuery() {
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorCtx(t.Robot.Ctx, "handleCommand panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	if t.Update.CallbackQuery != nil && t.Update.CallbackQuery.Message.ReplyToMessage != nil {
		t.Update.CallbackQuery.Message.MessageID = t.Update.CallbackQuery.Message.ReplyToMessage.MessageID
		t.Prompt = t.Update.CallbackQuery.Data
		cmd := strings.TrimSpace(strings.ReplaceAll(t.Update.CallbackQuery.Message.ReplyToMessage.Text, "@"+t.Bot.Self.UserName, ""))
		switch cmd {
		case "/" + param.TxtType, "/" + param.PhotoType, "/" + param.VideoType, "/" + param.RecType, "/" + param.TtsType:
			t.Robot.changeType(cmd)
			return
		case "/" + param.TxtModel, "/" + param.PhotoModel, "/" + param.VideoModel, "/" + param.RecModel, "/" + param.TtsModel:
			t.Robot.changeModel(cmd)
			return
		}
	}
	
	t.Robot.ExecCmd(t.Update.CallbackQuery.Data, nil, nil, nil)
}

// sendVideo send video to telegram
func (t *TelegramRobot) sendVideo() {
	t.Robot.TalkingPreCheck(func() {
		chatId, replyToMessageID, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := t.Prompt
		if len(prompt) == 0 {
			err := ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(replyToMessageID), "video_empty_content", t.Bot)
			if err != nil {
				logger.WarnCtx(t.Robot.Ctx, "force reply fail", "err", err)
			}
			return
		}
		
		imageContent := t.ImageContent
		thinkingMsgId := t.Robot.SendMsg(chatId, i18n.GetMessage("thinking", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		
		videoContent, totalToken, err := t.Robot.CreateVideo(prompt, imageContent)
		
		video := tgbotapi.NewInputMediaVideo(tgbotapi.FileBytes{
			Name:  "video." + utils.DetectVideoMimeType(videoContent),
			Bytes: videoContent,
		})
		
		edit := tgbotapi.EditMessageMediaConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:    int64(utils.ParseInt(chatId)),
				MessageID: utils.ParseInt(thinkingMsgId),
			},
			Media: video,
		}
		
		_, err = t.Bot.Request(edit)
		if err != nil {
			logger.WarnCtx(t.Robot.Ctx, "send video fail", "result", edit)
			t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			return
		}
		
		t.Robot.saveRecord(videoContent, imageContent, param.VideoRecordType, totalToken)
	})
}

// sendImg send img to telegram
func (t *TelegramRobot) sendImg() {
	t.Robot.TalkingPreCheck(func() {
		chatId, replyToMessageID, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
		
		prompt := t.Prompt
		if prompt == "" && t.Update.Message != nil && len(t.Update.Message.Photo) > 0 {
			prompt = t.Update.Message.Caption
		}
		
		var err error
		if len(prompt) == 0 {
			if strings.Contains(t.Command, "edit_photo") {
				err = ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(replyToMessageID), "edit_photo_empty_content", t.Bot)
			} else {
				err = ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(replyToMessageID), "photo_empty_content", t.Bot)
			}
			
			if err != nil {
				logger.WarnCtx(t.Robot.Ctx, "force reply fail", "err", err)
			}
			return
		}
		
		lastImageContent := t.ImageContent
		if len(lastImageContent) == 0 && strings.Contains(t.Command, "edit_photo") {
			lastImageContent, err = t.Robot.GetLastImageContent()
			if err != nil {
				logger.WarnCtx(t.Robot.Ctx, "get last image record fail", "err", err)
			}
		}
		
		thinkingMsgId := t.Robot.SendMsg(chatId, i18n.GetMessage("thinking", nil),
			replyToMessageID, tgbotapi.ModeMarkdown, nil)
		
		var photo tgbotapi.InputMediaPhoto
		imageContent, totalToken, err := t.Robot.CreatePhoto(prompt, lastImageContent)
		if err != nil {
			logger.WarnCtx(t.Robot.Ctx, "generate image fail", "err", err)
			t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			return
		}
		
		img, _, err := image.Decode(bytes.NewReader(imageContent))
		if err != nil {
			logger.ErrorCtx(t.Robot.Ctx, "decode image fail", "err", err)
			return
		}
		resizedImg := imaging.Fit(img, 1280, 1280, imaging.Lanczos)
		var buf bytes.Buffer
		
		err = imaging.Encode(&buf, resizedImg, imaging.JPEG)
		if err != nil {
			logger.ErrorCtx(t.Robot.Ctx, "encode image fail", "err", err)
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
			logger.WarnCtx(t.Robot.Ctx, "send image fail", "result", edit)
			t.Robot.SendMsg(chatId, err.Error(), replyToMessageID, "", nil)
			return
		}
		
		t.Robot.saveRecord(imageContent, lastImageContent, param.ImageRecordType, totalToken)
	})
}

// ExecuteForceReply use force reply interact with user
func (t *TelegramRobot) ExecuteForceReply() {
	defer func() {
		if err := recover(); err != nil {
			logger.ErrorCtx(t.Robot.Ctx, "ExecuteForceReply panic err", "err", err, "stack", string(debug.Stack()))
		}
	}()
	
	switch t.getMessage().ReplyToMessage.Text {
	case i18n.GetMessage("chat_empty_content", nil):
		t.sendChatMessage()
	case i18n.GetMessage("photo_empty_content", nil):
		t.sendImg()
	case i18n.GetMessage("edit_photo_empty_content", nil):
		t.Command = param.EditPhoto
		t.sendImg()
	case i18n.GetMessage("video_empty_content", nil):
		t.sendVideo()
	case i18n.GetMessage("task_empty_content", nil):
		t.Robot.sendMultiAgent("task_empty_content", t.sendForceReply("task_empty_content"))
	case i18n.GetMessage("mcp_empty_content", nil):
		t.Robot.sendMultiAgent("task_empty_content", t.sendForceReply("mcp_empty_content"))
	}
}

func (t *TelegramRobot) getContent(content string) (string, error) {
	var err error
	if t.Update.Message.Photo != nil {
		content, err = t.Robot.GetImageContent(t.ImageContent, content)
		if err != nil {
			logger.WarnCtx(t.Robot.Ctx, "get image content err", "err", err)
			return "", err
		}
	}
	
	if content == "" {
		logger.WarnCtx(t.Robot.Ctx, "content empty")
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
		logger.WarnCtx(t.Robot.Ctx, "get file fail", "err", err)
		return nil
	}
	
	downloadURL := file.Link(t.Bot.Token)
	voice, err := utils.DownloadFile(downloadURL)
	if err != nil {
		logger.WarnCtx(t.Robot.Ctx, "read response fail", "err", err)
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
		logger.WarnCtx(t.Robot.Ctx, "get file fail", "err", err)
		return nil
	}
	
	downloadURL := file.Link(t.Bot.Token)
	photoContent, err := utils.DownloadFile(downloadURL)
	if err != nil {
		logger.WarnCtx(t.Robot.Ctx, "read response fail", "err", err)
		return nil
	}
	
	return photoContent
}

func (t *TelegramRobot) getMessage() *tgbotapi.Message {
	if t.Update.Message != nil {
		return t.Update.Message
	}
	if t.Update.CallbackQuery != nil {
		return t.Update.CallbackQuery.Message
	}
	if t.Update.EditedMessage != nil {
		return t.Update.EditedMessage
	}
	return nil
}

func (t *TelegramRobot) getPrompt() string {
	return t.Prompt
}

func (t *TelegramRobot) sendForceReply(agentType string) func() {
	return func() {
		chatId, msgId, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
		err := ForceReply(int64(utils.ParseInt(chatId)), utils.ParseInt(msgId), agentType, t.Bot)
		if err != nil {
			logger.WarnCtx(t.Robot.Ctx, "force reply fail", "err", err)
		}
	}
}

func (t *TelegramRobot) getPerMsgLen() int {
	return 3596
}

func ForceReply(chatId int64, msgId int, i18MsgId string, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatId, i18n.GetMessage(i18MsgId, nil))
	msg.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	msg.ReplyToMessageID = msgId
	_, err := bot.Send(msg)
	return err
}

func (t *TelegramRobot) sendText(messageChan *MsgChan) {
	var msg *param.MsgInfo
	chatIdStr, msgIdStr, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	msgId := utils.ParseInt(msgIdStr)
	chatId := int64(utils.ParseInt(chatIdStr))
	parseMode := tgbotapi.ModeMarkdown
	
	for msg = range messageChan.NormalMessageChan {
		if len(msg.Content) == 0 {
			msg.Content = "get nothing from llm!"
		}
		
		if msg.MsgId == "" {
			tgMsgInfo := tgbotapi.NewMessage(chatId, msg.Content)
			tgMsgInfo.ReplyToMessageID = msgId
			tgMsgInfo.ParseMode = parseMode
			sendInfo, err := t.Bot.Send(tgMsgInfo)
			if err != nil {
				if t.sleepUtilNoLimit(msgId, err) {
					sendInfo, err = t.Bot.Send(tgMsgInfo)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					tgMsgInfo.ParseMode = ""
					sendInfo, err = t.Bot.Send(tgMsgInfo)
				} else {
					_, err = t.Bot.Send(tgMsgInfo)
				}
				if err != nil {
					logger.WarnCtx(t.Robot.Ctx, "Error sending message:", "msgID", msgId, "err", err)
					continue
				}
			}
			msg.MsgId = strconv.Itoa(sendInfo.MessageID)
		} else {
			updateMsg := tgbotapi.NewEditMessageText(chatId, utils.ParseInt(msg.MsgId), msg.Content)
			updateMsg.ParseMode = parseMode
			_, err := t.Bot.Send(updateMsg)
			if err != nil {
				// try again
				if t.sleepUtilNoLimit(msgId, err) {
					_, err = t.Bot.Send(updateMsg)
				} else if strings.Contains(err.Error(), "can't parse entities") {
					updateMsg.ParseMode = ""
					_, err = t.Bot.Send(updateMsg)
				} else {
					_, err = t.Bot.Send(updateMsg)
				}
				if err != nil {
					logger.WarnCtx(t.Robot.Ctx, "Error editing message", "msgID", msgId, "err", err)
				}
			}
		}
		
	}
}

func (t *TelegramRobot) sendVoiceContent(voiceContent []byte, duration int) error {
	chatIdStr, _, _ := t.Robot.GetChatIdAndMsgIdAndUserID()
	chatId := int64(utils.ParseInt(chatIdStr))
	
	_, err := t.Bot.Send(tgbotapi.NewVoice(chatId, tgbotapi.FileBytes{
		Name:  "voice." + utils.DetectAudioFormat(voiceContent),
		Bytes: voiceContent,
	}))
	if err != nil {
		logger.WarnCtx(t.Robot.Ctx, "send voice fail", "err", err)
		return err
	}
	
	return nil
}

func (t *TelegramRobot) setCommand(command string) {
	t.Command = command
}

func (t *TelegramRobot) getCommand() string {
	return t.Command
}

func (t *TelegramRobot) getUserName() string {
	return t.UserName
}

func (t *TelegramRobot) setPrompt(prompt string) {
	t.Prompt = prompt
}

func (t *TelegramRobot) getAudio() []byte {
	return t.AudioContent
}

func (t *TelegramRobot) getImage() []byte {
	return t.ImageContent
}
