package robot

import (
	"fmt"
	"strconv"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type Robot struct {
	TelegramRobot *TelegramRobot
	DiscordRobot  *DiscordRobot
}

type botOption func(r *Robot)

func NewRobot(options ...botOption) *Robot {
	r := new(Robot)
	for _, o := range options {
		o(r)
	}
	return r
}

func (r *Robot) Exec() {
	chatId, msgId, userId := r.GetChatIdAndMsgIdAndUserID()
	
	if !r.checkUserAllow(userId) && !r.checkGroupAllow(chatId) {
		logger.Warn("user/group not allow to use this bot", "userID", userId, "chat", chatId)
		r.SendMsg(chatId, i18n.GetMessage(*conf.BaseConfInfo.Lang, "valid_user_group", nil),
			msgId, tgbotapi.ModeMarkdown, nil)
		return
	}
	
	switch {
	case r.TelegramRobot != nil:
		r.TelegramRobot.Execute()
	case r.DiscordRobot != nil:
		r.DiscordRobot.Execute()
	}
}

func (r *Robot) GetChatIdAndMsgIdAndUserID() (int64, int, int64) {
	chatId := int64(0)
	msgId := 0
	userId := int64(0)
	
	switch {
	case r.TelegramRobot != nil:
		if r.TelegramRobot.Update.Message != nil {
			chatId = r.TelegramRobot.Update.Message.Chat.ID
			userId = r.TelegramRobot.Update.Message.From.ID
			msgId = r.TelegramRobot.Update.Message.MessageID
		}
		if r.TelegramRobot.Update.CallbackQuery != nil {
			chatId = r.TelegramRobot.Update.CallbackQuery.Message.Chat.ID
			userId = r.TelegramRobot.Update.CallbackQuery.From.ID
			msgId = r.TelegramRobot.Update.CallbackQuery.Message.MessageID
		}
	case r.DiscordRobot != nil:
		if r.DiscordRobot.Msg != nil {
			chatId, _ = strconv.ParseInt(r.DiscordRobot.Msg.ChannelID, 10, 64)
			userId, _ = strconv.ParseInt(r.DiscordRobot.Msg.Author.ID, 10, 64)
			msgId = utils.ParseInt(r.DiscordRobot.Msg.Message.ID)
		}
	}
	
	return chatId, msgId, userId
}

func (r *Robot) SendMsg(chatId int64, msgContent string, replyToMessageID int,
	parseMode string, inlineKeyboard *tgbotapi.InlineKeyboardMarkup) int {
	switch {
	case r.TelegramRobot != nil:
		msg := tgbotapi.NewMessage(chatId, msgContent)
		msg.ParseMode = parseMode
		msg.ReplyMarkup = inlineKeyboard
		msg.ReplyToMessageID = replyToMessageID
		msgInfo, err := r.TelegramRobot.Bot.Send(msg)
		if err != nil {
			logger.Warn("send clear message fail", "err", err)
			return 0
		}
		return msgInfo.MessageID
	case r.DiscordRobot != nil:
		//SendMsgToDiscord(chatId, msgContent, r.DiscordRobot, replyToMessageID, parseMode)
	}
	
	return 0
}

func WithTelegramRobot(TelegramBot *TelegramRobot) func(*Robot) {
	return func(r *Robot) {
		r.TelegramRobot = TelegramBot
	}
}

func WithDiscordRobot(discordBot *DiscordRobot) func(*Robot) {
	return func(r *Robot) {
		r.DiscordRobot = discordBot
	}
}

func StartRobot() {
	if *conf.BaseConfInfo.TelegramBotToken != "" {
		go func() {
			StartTelegramRobot()
		}()
	}
	
	if *conf.BaseConfInfo.DiscordBotToken != "" {
		go func() {
			StartDiscordRobot()
		}()
	}
}

// checkUserAllow check use can use telegram bot or not
func (r *Robot) checkUserAllow(userId int64) bool {
	if len(conf.BaseConfInfo.AllowedTelegramUserIds) == 0 {
		return true
	}
	if conf.BaseConfInfo.AllowedTelegramUserIds[0] {
		return false
	}
	
	_, ok := conf.BaseConfInfo.AllowedTelegramUserIds[userId]
	return ok
}

func (r *Robot) checkGroupAllow(chatId int64) bool {
	
	if len(conf.BaseConfInfo.AllowedTelegramGroupIds) == 0 {
		return true
	}
	if conf.BaseConfInfo.AllowedTelegramGroupIds[0] {
		return false
	}
	if _, ok := conf.BaseConfInfo.AllowedTelegramGroupIds[chatId]; ok {
		return true
	}
	
	return false
}

// checkUserTokenExceed check use token exceeded
func (r *Robot) checkUserTokenExceed(chatId int64, msgId int, userId int64) bool {
	if *conf.BaseConfInfo.TokenPerUser == 0 {
		return false
	}
	
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
		r.SendMsg(chatId, content, msgId, tgbotapi.ModeMarkdown, nil)
		return true
	}
	
	return false
}

// checkAdminUser check user is admin
func (r *Robot) checkAdminUser(userId int64) bool {
	if len(conf.BaseConfInfo.AdminUserIds) == 0 {
		return false
	}
	
	_, ok := conf.BaseConfInfo.AdminUserIds[userId]
	return ok
}
