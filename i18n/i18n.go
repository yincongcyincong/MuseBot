package i18n

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	botUtils "github.com/yincongcyincong/MuseBot/utils"
	"golang.org/x/text/language"
)

var (
	ruLocalizer *i18n.Localizer
	enLocalizer *i18n.Localizer
	zhLocalizer *i18n.Localizer
)

const (
	ru = "ru"
	en = "en"
	zh = "zh"
)

func InitI18n() {
	// 1. Create a new i18n bundle with English as default language
	bundle := i18n.NewBundle(language.English)

	// 2. Register JSON unmarshal function (other formats like TOML are also supported)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// 3. Load translation files
	// Russian translations
	if _, err := bundle.LoadMessageFile(botUtils.GetAbsPath("conf/i18n/i18n.ru.json")); err != nil {
		logger.Error("Failed to load Russian translation file", "err", err)
	}
	// English translations
	if _, err := bundle.LoadMessageFile(botUtils.GetAbsPath("conf/i18n/i18n.en.json")); err != nil {
		logger.Error("Failed to load English translation file", "err", err)
	}
	// Chinese translations
	if _, err := bundle.LoadMessageFile(botUtils.GetAbsPath("conf/i18n/i18n.zh.json")); err != nil {
		logger.Error("Failed to load Chinese translation file", "err", err)
	}

	// 4. Create localizers for each language
	ruLocalizer = i18n.NewLocalizer(bundle, ru)
	enLocalizer = i18n.NewLocalizer(bundle, en)
	zhLocalizer = i18n.NewLocalizer(bundle, zh)
}

// GetMessage function to get localized message
func GetMessage(tag string, messageID string, templateData map[string]interface{}) string {
	var localizer *i18n.Localizer
	switch tag {
	case ru:
		localizer = ruLocalizer
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
