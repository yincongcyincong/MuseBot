package utils

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func GetChatIdAndMsgIdAndUserID(update tgbotapi.Update) (int64, int, int64) {
	chatId := int64(0)
	msgId := 0
	userId := int64(0)
	if update.Message != nil {
		chatId = update.Message.Chat.ID
		userId = update.Message.From.ID
		msgId = update.Message.MessageID
	}
	if update.CallbackQuery != nil {
		chatId = update.CallbackQuery.Message.Chat.ID
		userId = update.CallbackQuery.From.ID
		msgId = update.CallbackQuery.Message.MessageID
	}

	return chatId, msgId, userId
}

func CheckMsgIsCallback(update tgbotapi.Update) bool {
	return update.CallbackQuery != nil
}
