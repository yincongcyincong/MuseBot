package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"sync"
)

var (
	userChatMap = sync.Map{}
)

func CheckUserChatExceed(update tgbotapi.Update, bot *tgbotapi.BotAPI) bool {
	chatId, msgId, userId := GetChatIdAndMsgIdAndUserID(update)
	times := 1
	if timeInter, ok := userChatMap.Load(userId); ok {
		times = timeInter.(int)
		if times >= *conf.MaxUserChat {
			i18n.SendMsg(chatId, "chat_exceed", bot, nil, msgId)
			return true
		}
		times++
	}
	userChatMap.Store(userId, times)
	return false
}

func DecreaseUserChat(update tgbotapi.Update) {
	_, _, userId := GetChatIdAndMsgIdAndUserID(update)
	if timeInter, ok := userChatMap.Load(userId); ok {
		times := timeInter.(int)
		times--
		userChatMap.Store(userId, times)
	}
}
