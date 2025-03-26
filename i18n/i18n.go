package i18n

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"golang.org/x/text/language"
)

var (
	enLocalizer *i18n.Localizer
	zhLocalizer *i18n.Localizer
)

const (
	en = "en"
	zh = "zh"
)

func InitI18n() {
	// 1. Create a new i18n bundle with English as default language
	bundle := i18n.NewBundle(language.English)

	// 2. Register JSON unmarshal function (other formats like TOML are also supported)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 3. Load translation files
	// English translations
	if _, err := bundle.LoadMessageFile("./conf/i18n.en.json"); err != nil {
		logger.Fatal("Failed to load English translation file", "err", err)
	}
	// Chinese translations
	if _, err := bundle.LoadMessageFile("./conf/i18n.zh.json"); err != nil {
		logger.Fatal("Failed to load Chinese translation file", "err", err)
	}

	// 4. Create localizers for each language
	enLocalizer = i18n.NewLocalizer(bundle, en)
	zhLocalizer = i18n.NewLocalizer(bundle, zh)
}

// GetMessage function to get localized message
func GetMessage(tag string, messageID string, templateData map[string]interface{}) string {
	var localizer *i18n.Localizer
	switch tag {
	case zh:
		localizer = zhLocalizer
	default:
		localizer = enLocalizer
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		logger.Warn("Failed to localize message", "tag", tag, "messageID", messageID, "err", err)
		return ""
	}
	return msg
}

// SendMsg send message to user
func SendMsg(chatId int64, msgId string, bot *tgbotapi.BotAPI, inlineKeyboard *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatId, GetMessage(*conf.Lang, msgId, nil))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = inlineKeyboard
	_, err := bot.Send(msg)
	if err != nil {
		logger.Warn("send clear message fail", "err", err)
	}
}
