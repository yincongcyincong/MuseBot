package utils

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func GetChatIdAndMsgIdAndUserName(update tgbotapi.Update) (int64, int, string) {
	chatId := int64(0)
	msgId := 0
	username := ""
	if update.Message != nil {
		chatId = update.Message.Chat.ID
		username = update.Message.From.String()
		msgId = update.Message.MessageID
	}
	if update.CallbackQuery != nil {
		chatId = update.CallbackQuery.Message.Chat.ID
		username = update.CallbackQuery.From.String()
		msgId = update.CallbackQuery.Message.MessageID
	}

	return chatId, msgId, username
}

func CheckMsgIsCallback(update tgbotapi.Update) bool {
	return update.CallbackQuery != nil
}
