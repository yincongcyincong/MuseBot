package utils

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetChatIdAndMsgIdAndUserName_MessageUpdate(t *testing.T) {
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: 123456},
			From:      &tgbotapi.User{UserName: "test_user"},
			MessageID: 789,
		},
	}

	chatId, msgId, username := GetChatIdAndMsgIdAndUserName(update)
	assert.Equal(t, int64(123456), chatId)
	assert.Equal(t, 789, msgId)
	assert.Equal(t, "test_user", username)
}

func TestGetChatIdAndMsgIdAndUserName_CallbackQueryUpdate(t *testing.T) {
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat:      &tgbotapi.Chat{ID: 654321},
				MessageID: 456,
			},
			From: &tgbotapi.User{UserName: "callback_user"},
		},
	}

	chatId, msgId, username := GetChatIdAndMsgIdAndUserName(update)
	assert.Equal(t, int64(654321), chatId)
	assert.Equal(t, 456, msgId)
	assert.Equal(t, "callback_user", username)
}

func TestGetChatIdAndMsgIdAndUserName_EmptyUpdate(t *testing.T) {
	update := tgbotapi.Update{}

	chatId, msgId, username := GetChatIdAndMsgIdAndUserName(update)
	assert.Equal(t, int64(0), chatId)
	assert.Equal(t, 0, msgId)
	assert.Equal(t, "", username)
}

func TestCheckMsgIsCallback(t *testing.T) {
	updateWithCallback := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{},
	}
	assert.True(t, CheckMsgIsCallback(updateWithCallback))

	updateWithMessage := tgbotapi.Update{
		Message: &tgbotapi.Message{},
	}
	assert.False(t, CheckMsgIsCallback(updateWithMessage))

	updateEmpty := tgbotapi.Update{}
	assert.False(t, CheckMsgIsCallback(updateEmpty))
}
