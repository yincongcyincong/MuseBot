package utils

import (
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

func GetChatType(update tgbotapi.Update) string {
	chat := GetChat(update)
	return chat.Type
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

func ReplaceCommand(content string, command string, botName string) string {
	mention := "@" + botName

	content = strings.ReplaceAll(content, command, mention)
	content = strings.ReplaceAll(content, mention, "")
	prompt := strings.TrimSpace(content)

	return prompt
}

func ForceReply(chatId int64, msgId int, i18MsgId string, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatId, i18n.GetMessage(*conf.Lang, i18MsgId, nil))
	msg.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	msg.ReplyToMessageID = msgId
	_, err := bot.Send(msg)
	return err
}

func GetAudioContent(update tgbotapi.Update, bot *tgbotapi.BotAPI) []byte {
	if update.Message == nil || update.Message.Voice == nil {
		return nil
	}

	fileID := update.Message.Voice.FileID
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		logger.Warn("get file fail", "err", err)
		return nil
	}

	// 构造下载 URL
	downloadURL := file.Link(bot.Token)

	transport := &http.Transport{}

	if *conf.TelegramProxy != "" {
		proxy, err := url.Parse(*conf.TelegramProxy)
		if err != nil {
			logger.Warn("parse proxy url fail", "err", err)
			return nil
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // 设置超时
	}

	// 通过代理下载
	resp, err := client.Get(downloadURL)
	if err != nil {
		logger.Warn("download fail", "err", err)
		return nil
	}
	defer resp.Body.Close()
	voice, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("read response fail", "err", err)
		return nil
	}
	return voice
}

func GetPhotoContent(update tgbotapi.Update, bot *tgbotapi.BotAPI) []byte {
	if update.Message == nil || update.Message.Photo == nil {
		return nil
	}

	var photo tgbotapi.PhotoSize
	for i := len(update.Message.Photo) - 1; i >= 0; i-- {
		if update.Message.Photo[i].FileSize < 8*1024*1024 {
			photo = update.Message.Photo[i]
			break
		}
	}

	fileID := photo.FileID
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		logger.Warn("get file fail", "err", err)
		return nil
	}

	// 构造下载 URL
	downloadURL := file.Link(bot.Token)

	transport := &http.Transport{}

	if *conf.TelegramProxy != "" {
		proxy, err := url.Parse(*conf.TelegramProxy)
		if err != nil {
			logger.Warn("parse proxy url fail", "err", err)
			return nil
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // 设置超时
	}

	// 通过代理下载
	resp, err := client.Get(downloadURL)
	if err != nil {
		logger.Warn("download fail", "err", err)
		return nil
	}
	defer resp.Body.Close()
	photoContent, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("read response fail", "err", err)
		return nil
	}

	return photoContent
}
