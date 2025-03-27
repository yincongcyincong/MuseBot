package utils

import (
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"strconv"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func GetChat(update tgbotapi.Update) *tgbotapi.Chat {
	if update.Message != nil {
		return update.Message.Chat
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat
	}
	return nil
}

func GetMessage(update tgbotapi.Update) *tgbotapi.Message {
	if update.Message != nil {
		return update.Message
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message
	}
	return nil
}

func CheckMsgIsCallback(update tgbotapi.Update) bool {
	return update.CallbackQuery != nil
}

// Utf16len calculates the length of a string in UTF-16 code units.
func Utf16len(s string) int {
	utf16Str := utf16.Encode([]rune(s))
	return len(utf16Str)
}

func ParseInt(str string) int {
	num, _ := strconv.Atoi(str)
	return num
}

func SendMsg(chatId int64, msgContent string, bot *tgbotapi.BotAPI, replyToMessageID int) {
	msg := tgbotapi.NewMessage(chatId, msgContent)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyToMessageID = replyToMessageID
	_, err := bot.Send(msg)
	if err != nil {
		logger.Warn("send clear message fail", "err", err)
	}
}
