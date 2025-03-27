package conf

import (
	"flag"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	BotToken      *string
	DeepseekToken *string
	DeepseekType  *string // simple complex
	CustomUrl     *string
	VolcAK        *string
	VolcSK        *string
	DBType        *string
	DBConf        *string
	DeepseekProxy *string
	TelegramProxy *string
	Lang          *string
	TokenPerUser  *int

	AllowedTelegramUserIds  = make(map[int64]bool)
	AllowedTelegramGroupIds = make(map[int64]bool)
	AdminUserIds            = make(map[int64]bool)
)

var (
	Bot *tgbotapi.BotAPI
)

func InitConf() {
	BotToken = flag.String("telegram_bot_token", "", "Comma-separated list of Telegram bot tokens")
	DeepseekToken = flag.String("deepseek_token", "", "deepseek auth token")
	CustomUrl = flag.String("custom_url", "https://api.deepseek.com/", "deepseek custom url")
	DeepseekType = flag.String("deepseek_type", "deepseek", "deepseek auth type")
	VolcAK = flag.String("volc_ak", "", "volc ak")
	VolcSK = flag.String("volc_sk", "", "volc sk")
	DBType = flag.String("db_type", "sqlite3", "db type")
	DBConf = flag.String("db_conf", "./data/telegram_bot.db", "db conf")
	DeepseekProxy = flag.String("deepseek_proxy", "", "db conf")
	TelegramProxy = flag.String("telegram_proxy", "", "db conf")
	Lang = flag.String("lang", "en", "lang")
	TokenPerUser = flag.Int("token_per_user", 10000, "token per user")

	adminUserIds := flag.String("admin_user_ids", "", "admin user ids")
	allowedUserIds := flag.String("allowed_telegram_user_ids", "", "allowed telegram user ids")
	allowedGroupIds := flag.String("allowed_telegram_group_ids", "", "allowed telegram group ids")
	flag.Parse()

	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" {
		*BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	if os.Getenv("DEEPSEEK_TOKEN") != "" {
		*DeepseekToken = os.Getenv("DEEPSEEK_TOKEN")
	}

	if os.Getenv("CUSTOM_URL") != "" {
		*CustomUrl = os.Getenv("CUSTOM_URL")
	}

	if os.Getenv("DEEPSEEK_TYPE") != "" {
		*DeepseekType = os.Getenv("DEEPSEEK_TYPE")
	}

	if os.Getenv("VOLC_AK") != "" {
		*VolcAK = os.Getenv("VOLC_AK")
	}

	if os.Getenv("VOLC_SK") != "" {
		*VolcSK = os.Getenv("VOLC_SK")
	}

	if os.Getenv("DB_TYPE") != "" {
		*DBType = os.Getenv("DB_TYPE")
	}

	if os.Getenv("DB_CONF") != "" {
		*DBConf = os.Getenv("DB_CONF")
	}

	if os.Getenv("ALLOWED_TELEGRAM_USER_IDS") != "" {
		*allowedUserIds = os.Getenv("ALLOWED_TELEGRAM_USER_IDS")
	}

	if os.Getenv("ALLOWED_TELEGRAM_GROUP_IDS") != "" {
		*allowedGroupIds = os.Getenv("ALLOWED_TELEGRAM_GROUP_IDS")
	}

	if os.Getenv("DEEPSEEK_PROXY") != "" {
		*DeepseekProxy = os.Getenv("DEEPSEEK_PROXY")
	}

	if os.Getenv("TELEGRAM_PROXY") != "" {
		*TelegramProxy = os.Getenv("TELEGRAM_PROXY")
	}

	if os.Getenv("LANG") != "" {
		*Lang = os.Getenv("LANG")
	}

	if os.Getenv("TOKEN_PER_USER") != "" {
		*TokenPerUser, _ = strconv.Atoi(os.Getenv("TOKEN_PER_USER"))
	}

	if os.Getenv("ADMIN_USER_IDS") != "" {
		*adminUserIds = os.Getenv("ADMIN_USER_IDS")
	}

	for _, userIdStr := range strings.Split(*allowedUserIds, ",") {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			logger.Warn("AllowedTelegramUserIds parse error", "userID", userIdStr)
			continue
		}
		AllowedTelegramUserIds[int64(userId)] = true
	}

	for _, groupIdStr := range strings.Split(*allowedGroupIds, ",") {
		groupId, err := strconv.Atoi(groupIdStr)
		if err != nil {
			logger.Warn("AllowedTelegramGroupIds parse error", "groupId", groupIdStr)
			continue
		}
		AllowedTelegramGroupIds[int64(groupId)] = true
	}

	for _, userIdStr := range strings.Split(*adminUserIds, ",") {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			logger.Warn("AdminUserIds parse error", "userID", userIdStr)
			continue
		}
		AdminUserIds[int64(userId)] = true
	}

	logger.Info("", "TelegramBotToken", *BotToken)
	logger.Info("", "DeepseekToken", *DeepseekToken)
	logger.Info("", "CustomUrl", *CustomUrl)
	logger.Info("", "DeepseekType", *DeepseekType)
	logger.Info("", "VOLC_AK", *VolcAK)
	logger.Info("", "VOLC_SK", *VolcSK)
	logger.Info("", "DBType", *DBType)
	logger.Info("", "DBConf", *DBConf)
	logger.Info("", "AllowedTelegramUserIds", *allowedUserIds)
	logger.Info("", "AllowedTelegramGroupIds", *allowedGroupIds)
	logger.Info("", "DeepseekProxy", *DeepseekProxy)
	logger.Info("", "TelegramProxy", *TelegramProxy)
	logger.Info("", "LANG", *Lang)
	logger.Info("", "TOKEN_PER_USER", *TokenPerUser)
	logger.Info("", "AdminUserIds", *adminUserIds)

	if *BotToken == "" || *DeepseekToken == "" {
		panic("Bot token and deepseek token are required")
	}

}

func CreateBot() *tgbotapi.BotAPI {
	// 配置自定义 HTTP Client 并设置代理
	client := &http.Client{}

	// parse proxy URL
	if *TelegramProxy != "" {
		proxy, err := url.Parse(*TelegramProxy)
		if err != nil {
			logger.Info("Failed to parse proxy URL", err)
		} else {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	var err error
	Bot, err = tgbotapi.NewBotAPIWithClient(*BotToken, tgbotapi.APIEndpoint, client)
	if err != nil {
		panic("Init bot fail" + err.Error())
	}

	if *logger.LogLevel == "debug" {
		Bot.Debug = true
	}

	// set command
	cmdCfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "help",
			Description: "help",
		},
		tgbotapi.BotCommand{
			Command:     "clear",
			Description: "clear all of your communication record with deepseek.",
		},
		tgbotapi.BotCommand{
			Command:     "retry",
			Description: "retry last question.",
		},
		tgbotapi.BotCommand{
			Command:     "mode",
			Description: "chose deepseek mode, include chat, coder, reasoner",
		},
		tgbotapi.BotCommand{
			Command:     "balance",
			Description: "show deepseek balance.",
		},
		tgbotapi.BotCommand{
			Command:     "state",
			Description: "calculate one user token usage.",
		},
		tgbotapi.BotCommand{
			Command:     "photo",
			Description: "using volcengine photo model create photo.",
		},
		tgbotapi.BotCommand{
			Command:     "video",
			Description: "using volcengine video model create video.",
		},
		tgbotapi.BotCommand{
			Command:     "chat",
			Description: "allows the bot to chat through /chat command in groups, without the bot being set as admin of the group.",
		},
	)
	Bot.Send(cmdCfg)

	return Bot
}
