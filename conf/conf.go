package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

var (
	BotToken        *string
	DeepseekToken   *string
	OpenAIToken     *string
	GeminiToken     *string
	OpenRouterToken *string
	VolToken        *string
	ErnieAK         *string
	ErnieSK         *string

	Type          *string // simple complex
	CustomUrl     *string
	VolcAK        *string
	VolcSK        *string
	DBType        *string
	DBConf        *string
	DeepseekProxy *string
	TelegramProxy *string
	Lang          *string
	TokenPerUser  *int
	NeedATBOt     *bool
	MaxUserChat   *int
	VideoToken    *string
	HTTPPort      *int
	UseTools      *bool

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
	OpenAIToken = flag.String("openai_token", "", "openai auth token")
	GeminiToken = flag.String("gemini_token", "", "gemini auth token")
	OpenRouterToken = flag.String("openrouter_token", "", "openrouter.ai auth token")
	VolToken = flag.String("vol_token", "", "vol auth token")
	ErnieAK = flag.String("ernie_ak", "", "ernie ak")
	ErnieSK = flag.String("ernie_sk", "", "ernie sk")
	VolcAK = flag.String("volc_ak", "", "volc ak")
	VolcSK = flag.String("volc_sk", "", "volc sk")

	CustomUrl = flag.String("custom_url", "https://api.deepseek.com/", "deepseek custom url")
	Type = flag.String("type", "deepseek", "llm type: deepseek gemini openai openrouter")
	DBType = flag.String("db_type", "sqlite3", "db type")
	DBConf = flag.String("db_conf", "./data/telegram_bot.db", "db conf")
	DeepseekProxy = flag.String("deepseek_proxy", "", "db conf")
	TelegramProxy = flag.String("telegram_proxy", "", "db conf")
	Lang = flag.String("lang", "en", "lang")
	TokenPerUser = flag.Int("token_per_user", 10000, "token per user")
	NeedATBOt = flag.Bool("need_at_bot", false, "need at bot")
	MaxUserChat = flag.Int("max_user_chat", 2, "max chat per user")
	VideoToken = flag.String("video_token", "", "video token")
	HTTPPort = flag.Int("http_port", 36060, "http server port")
	UseTools = flag.Bool("use_tools", true, "use tools")

	adminUserIds := flag.String("admin_user_ids", "", "admin user ids")
	allowedUserIds := flag.String("allowed_telegram_user_ids", "", "allowed telegram user ids")
	allowedGroupIds := flag.String("allowed_telegram_group_ids", "", "allowed telegram group ids")

	InitDeepseekConf()
	InitPhotoConf()
	InitVideoConf()
	InitAudioConf()
	InitToolsConf()
	InitRagConf()
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

	if os.Getenv("TYPE") != "" {
		*Type = os.Getenv("TYPE")
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

	if os.Getenv("NEED_AT_BOT") != "" {
		*NeedATBOt, _ = strconv.ParseBool(os.Getenv("NEED_AT_BOT"))
	}

	if os.Getenv("MAX_USER_CHAT") != "" {
		*MaxUserChat, _ = strconv.Atoi(os.Getenv("MAX_USER_CHAT"))
	}

	if os.Getenv("VIDEO_TOKEN") != "" {
		*VideoToken = os.Getenv("VIDEO_TOKEN")
	}

	if os.Getenv("HTTP_PORT") != "" {
		*HTTPPort, _ = strconv.Atoi(os.Getenv("HTTP_PORT"))
	}

	if os.Getenv("USE_TOOLS") == "false" {
		*UseTools = false
	}

	if os.Getenv("OPENAI_TOKEN") != "" {
		*OpenAIToken = os.Getenv("OPENAI_TOKEN")
	}

	if os.Getenv("GEMINI_TOKEN") != "" {
		*GeminiToken = os.Getenv("GEMINI_TOKEN")
	}

	if os.Getenv("VOL_TOKEN") != "" {
		*VolToken = os.Getenv("VOL_TOKEN")
	}

	if os.Getenv("ERNIE_AK") != "" {
		*ErnieAK = os.Getenv("ERNIE_AK")
	}

	if os.Getenv("ERNIE_SK") != "" {
		*ErnieSK = os.Getenv("ERNIE_SK")
	}

	if os.Getenv("OPEN_ROUTER_TOKEN") != "" {
		*OpenRouterToken = os.Getenv("OPEN_ROUTER_TOKEN")
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

	logger.Info("CONF", "TelegramBotToken", *BotToken)
	logger.Info("CONF", "DeepseekToken", *DeepseekToken)
	logger.Info("CONF", "CustomUrl", *CustomUrl)
	logger.Info("CONF", "Type", *Type)
	logger.Info("CONF", "VolcAK", *VolcAK)
	logger.Info("CONF", "VolcSK", *VolcSK)
	logger.Info("CONF", "DBType", *DBType)
	logger.Info("CONF", "DBConf", *DBConf)
	logger.Info("CONF", "AllowedTelegramUserIds", *allowedUserIds)
	logger.Info("CONF", "AllowedTelegramGroupIds", *allowedGroupIds)
	logger.Info("CONF", "DeepseekProxy", *DeepseekProxy)
	logger.Info("CONF", "TelegramProxy", *TelegramProxy)
	logger.Info("CONF", "Lang", *Lang)
	logger.Info("CONF", "TokenPerUser", *TokenPerUser)
	logger.Info("CONF", "AdminUserIds", *adminUserIds)
	logger.Info("CONF", "NeedATBOt", *NeedATBOt)
	logger.Info("CONF", "MaxUserChat", *MaxUserChat)
	logger.Info("CONF", "VideoToken", *VideoToken)
	logger.Info("CONF", "HTTPPort", *HTTPPort)
	logger.Info("CONF", "OpenAIToken", *OpenAIToken)
	logger.Info("CONF", "GeminiToken", *GeminiToken)
	logger.Info("CONF", "OpenRouterToken", OpenRouterToken)
	logger.Info("CONF", "ErnieAK", *ErnieAK)
	logger.Info("CONF", "ErnieSK", *ErnieSK)
	logger.Info("CONF", "VolToken", *VolToken)

	EnvAudioConf()
	EnvRagConf()
	EnvDeepseekConf()
	EnvPhotoConf()
	EnvToolsConf()
	EnvVideoConf()

	if *BotToken == "" {
		panic("Bot token and llm token are required")
	}

}
