package utils

import (
	"sync"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
)

var (
	userChatMap = sync.Map{}
)

func CheckUserChatExceed(userId string) bool {
	times := 1
	if timeInter, ok := userChatMap.Load(userId); ok {
		times = timeInter.(int)
		if times >= *conf.BaseConfInfo.MaxUserChat {
			return true
		}
		times++
	}
	userChatMap.Store(userId, times)
	return false
}

func DecreaseUserChat(userId string) {
	if timeInter, ok := userChatMap.Load(userId); ok {
		times := timeInter.(int)
		times--
		userChatMap.Store(userId, times)
	}
}
