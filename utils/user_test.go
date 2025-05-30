package utils

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestDecreaseUserChat(t *testing.T) {
	userId := int64(999999999)
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			MessageID: 1,

			From: &tgbotapi.User{
				ID: userId,
			},
			Chat: &tgbotapi.Chat{
				ID: 1,
			},
		},
	}

	// 初始化次数为 3
	userChatMap.Store(userId, 3)

	DecreaseUserChat(update)

	if val, ok := userChatMap.Load(userId); !ok || val.(int) != 2 {
		t.Errorf("Expected times to be 2, got %v", val)
	}
}
