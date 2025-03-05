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
			From:      &tgbotapi.User{ID: 123},
			MessageID: 789,
		},
	}

	chatId, msgId, userId := GetChatIdAndMsgIdAndUserID(update)
	assert.Equal(t, int64(123456), chatId)
	assert.Equal(t, 789, msgId)
	assert.Equal(t, 123, userId)
}

func TestGetChatIdAndMsgIdAndUserName_CallbackQueryUpdate(t *testing.T) {
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat:      &tgbotapi.Chat{ID: 654321},
				MessageID: 456,
			},
			From: &tgbotapi.User{ID: 111},
		},
	}

	chatId, msgId, userId := GetChatIdAndMsgIdAndUserID(update)
	assert.Equal(t, int64(654321), chatId)
	assert.Equal(t, 456, msgId)
	assert.Equal(t, 111, userId)
}

func TestGetChatIdAndMsgIdAndUserName_EmptyUpdate(t *testing.T) {
	update := tgbotapi.Update{}

	chatId, msgId, userId := GetChatIdAndMsgIdAndUserID(update)
	assert.Equal(t, int64(0), chatId)
	assert.Equal(t, 0, msgId)
	assert.Equal(t, 0, userId)
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
