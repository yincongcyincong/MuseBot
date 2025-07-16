package robot

import (
	"sync"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
)

var (
	userChatMap = sync.Map{}
)

func CheckUserChatExceed(userId int64) bool {
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

func DecreaseUserChat(userId int64) {
	if timeInter, ok := userChatMap.Load(userId); ok {
		times := timeInter.(int)
		times--
		userChatMap.Store(userId, times)
	}
}
