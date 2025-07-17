package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type BaseConf struct {
	TelegramBotToken *string `json:"telegram_bot_token"`
	DiscordBotToken  *string `json:"discord_bot_token"`
	DeepseekToken    *string `json:"deepseek_token"`
	OpenAIToken      *string `json:"openai_token"`
	GeminiToken      *string `json:"gemini_token"`
	OpenRouterToken  *string `json:"openrouter_token"`
	VolToken         *string `json:"vol_token"`
	ErnieAK          *string `json:"ernie_ak"`
	ErnieSK          *string `json:"ernie_sk"`
	
	Type          *string `json:"type"`
	MediaType     *string `json:"media_type"`
	CustomUrl     *string `json:"custom_url"`
	VolcAK        *string `json:"volc_ak"`
	VolcSK        *string `json:"volc_sk"`
	DBType        *string `json:"db_type"`
	DBConf        *string `json:"db_conf"`
	DeepseekProxy *string `json:"deepseek_proxy"`
	TelegramProxy *string `json:"telegram_proxy"`
	Lang          *string `json:"lang"`
	TokenPerUser  *int    `json:"token_per_user"`
	NeedATBOt     *bool   `json:"need_at_bot"`
	MaxUserChat   *int    `json:"max_user_chat"`
	VideoToken    *string `json:"video_token"`
	HTTPPort      *int    `json:"http_port"`
	UseTools      *bool   `json:"use_tools"`
	
	CrtFile *string `json:"crt_file"`
	KeyFile *string `json:"key_file"`
	CaFile  *string `json:"ca_file"`
	
	AllowedTelegramUserIds  map[string]bool `json:"allowed_telegram_user_ids"`
	AllowedTelegramGroupIds map[int64]bool  `json:"allowed_telegram_group_ids"`
	AdminUserIds            map[string]bool `json:"admin_user_ids"`
	
	Bot *tgbotapi.BotAPI `json:"bot"`
}

var (
	BaseConfInfo = new(BaseConf)
)

func InitConf() {
	BaseConfInfo.TelegramBotToken = flag.String("telegram_bot_token", "", "Telegram bot tokens")
	BaseConfInfo.DiscordBotToken = flag.String("discord_bot_token", "", "Discord bot tokens")
	BaseConfInfo.DeepseekToken = flag.String("deepseek_token", "", "deepseek auth token")
	BaseConfInfo.OpenAIToken = flag.String("openai_token", "", "openai auth token")
	BaseConfInfo.GeminiToken = flag.String("gemini_token", "", "gemini auth token")
	BaseConfInfo.OpenRouterToken = flag.String("openrouter_token", "", "openrouter.ai auth token")
	BaseConfInfo.VolToken = flag.String("vol_token", "", "vol auth token")
	BaseConfInfo.ErnieAK = flag.String("ernie_ak", "", "ernie ak")
	BaseConfInfo.ErnieSK = flag.String("ernie_sk", "", "ernie sk")
	BaseConfInfo.VolcAK = flag.String("volc_ak", "", "volc ak")
	BaseConfInfo.VolcSK = flag.String("volc_sk", "", "volc sk")
	
	BaseConfInfo.CustomUrl = flag.String("custom_url", "", "deepseek custom url")
	BaseConfInfo.Type = flag.String("type", "deepseek", "llm type: deepseek gemini openai openrouter vol")
	BaseConfInfo.MediaType = flag.String("media_type", "vol", "media type: vol gemini openai openrouter")
	BaseConfInfo.DBType = flag.String("db_type", "sqlite3", "db type")
	BaseConfInfo.DBConf = flag.String("db_conf", "./data/telegram_bot.db", "db conf")
	BaseConfInfo.DeepseekProxy = flag.String("deepseek_proxy", "", "db conf")
	BaseConfInfo.TelegramProxy = flag.String("telegram_proxy", "", "db conf")
	BaseConfInfo.Lang = flag.String("lang", "en", "lang")
	BaseConfInfo.TokenPerUser = flag.Int("token_per_user", 10000, "token per user")
	BaseConfInfo.NeedATBOt = flag.Bool("need_at_bot", false, "need at bot")
	BaseConfInfo.MaxUserChat = flag.Int("max_user_chat", 2, "max chat per user")
	BaseConfInfo.VideoToken = flag.String("video_token", "", "video token")
	BaseConfInfo.HTTPPort = flag.Int("http_port", 36060, "http server port")
	BaseConfInfo.UseTools = flag.Bool("use_tools", true, "use tools")
	
	BaseConfInfo.CrtFile = flag.String("crt_file", "", "public key file")
	BaseConfInfo.KeyFile = flag.String("key_file", "", "secret key file")
	BaseConfInfo.CaFile = flag.String("ca_file", "", "ca file")
	
	adminUserIds := flag.String("admin_user_ids", "", "admin user ids")
	allowedUserIds := flag.String("allowed_telegram_user_ids", "", "allowed telegram user ids")
	allowedGroupIds := flag.String("allowed_telegram_group_ids", "", "allowed telegram group ids")
	
	BaseConfInfo.AllowedTelegramUserIds = make(map[string]bool)
	BaseConfInfo.AllowedTelegramGroupIds = make(map[int64]bool)
	BaseConfInfo.AdminUserIds = make(map[string]bool)
	
	InitDeepseekConf()
	InitPhotoConf()
	InitVideoConf()
	InitAudioConf()
	InitToolsConf()
	InitRagConf()
	flag.Parse()
	
	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" {
		*BaseConfInfo.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	
	if os.Getenv("DISCORD_BOT_TOKEN") != "" {
		*BaseConfInfo.DiscordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	}
	
	if os.Getenv("DEEPSEEK_TOKEN") != "" {
		*BaseConfInfo.DeepseekToken = os.Getenv("DEEPSEEK_TOKEN")
	}
	
	if os.Getenv("CUSTOM_URL") != "" {
		*BaseConfInfo.CustomUrl = os.Getenv("CUSTOM_URL")
	}
	
	if os.Getenv("TYPE") != "" {
		*BaseConfInfo.Type = os.Getenv("TYPE")
	}
	
	if os.Getenv("VOLC_AK") != "" {
		*BaseConfInfo.VolcAK = os.Getenv("VOLC_AK")
	}
	
	if os.Getenv("VOLC_SK") != "" {
		*BaseConfInfo.VolcSK = os.Getenv("VOLC_SK")
	}
	
	if os.Getenv("DB_TYPE") != "" {
		*BaseConfInfo.DBType = os.Getenv("DB_TYPE")
	}
	
	if os.Getenv("DB_CONF") != "" {
		*BaseConfInfo.DBConf = os.Getenv("DB_CONF")
	}
	
	if os.Getenv("ALLOWED_TELEGRAM_USER_IDS") != "" {
		*allowedUserIds = os.Getenv("ALLOWED_TELEGRAM_USER_IDS")
	}
	
	if os.Getenv("ALLOWED_TELEGRAM_GROUP_IDS") != "" {
		*allowedGroupIds = os.Getenv("ALLOWED_TELEGRAM_GROUP_IDS")
	}
	
	if os.Getenv("DEEPSEEK_PROXY") != "" {
		*BaseConfInfo.DeepseekProxy = os.Getenv("DEEPSEEK_PROXY")
	}
	
	if os.Getenv("TELEGRAM_PROXY") != "" {
		*BaseConfInfo.TelegramProxy = os.Getenv("TELEGRAM_PROXY")
	}
	
	if os.Getenv("LANG") != "" {
		*BaseConfInfo.Lang = os.Getenv("LANG")
	}
	
	if os.Getenv("TOKEN_PER_USER") != "" {
		*BaseConfInfo.TokenPerUser, _ = strconv.Atoi(os.Getenv("TOKEN_PER_USER"))
	}
	
	if os.Getenv("ADMIN_USER_IDS") != "" {
		*adminUserIds = os.Getenv("ADMIN_USER_IDS")
	}
	
	if os.Getenv("NEED_AT_BOT") != "" {
		*BaseConfInfo.NeedATBOt, _ = strconv.ParseBool(os.Getenv("NEED_AT_BOT"))
	}
	
	if os.Getenv("MAX_USER_CHAT") != "" {
		*BaseConfInfo.MaxUserChat, _ = strconv.Atoi(os.Getenv("MAX_USER_CHAT"))
	}
	
	if os.Getenv("VIDEO_TOKEN") != "" {
		*BaseConfInfo.VideoToken = os.Getenv("VIDEO_TOKEN")
	}
	
	if os.Getenv("HTTP_PORT") != "" {
		*BaseConfInfo.HTTPPort, _ = strconv.Atoi(os.Getenv("HTTP_PORT"))
	}
	
	if os.Getenv("USE_TOOLS") == "false" {
		*BaseConfInfo.UseTools = false
	}
	
	if os.Getenv("OPENAI_TOKEN") != "" {
		*BaseConfInfo.OpenAIToken = os.Getenv("OPENAI_TOKEN")
	}
	
	if os.Getenv("GEMINI_TOKEN") != "" {
		*BaseConfInfo.GeminiToken = os.Getenv("GEMINI_TOKEN")
	}
	
	if os.Getenv("VOL_TOKEN") != "" {
		*BaseConfInfo.VolToken = os.Getenv("VOL_TOKEN")
	}
	
	if os.Getenv("ERNIE_AK") != "" {
		*BaseConfInfo.ErnieAK = os.Getenv("ERNIE_AK")
	}
	
	if os.Getenv("ERNIE_SK") != "" {
		*BaseConfInfo.ErnieSK = os.Getenv("ERNIE_SK")
	}
	
	if os.Getenv("OPEN_ROUTER_TOKEN") != "" {
		*BaseConfInfo.OpenRouterToken = os.Getenv("OPEN_ROUTER_TOKEN")
	}
	
	if os.Getenv("CRT_FILE") != "" {
		*BaseConfInfo.CrtFile = os.Getenv("CRT_FILE")
	}
	
	if os.Getenv("KEY_FILE") != "" {
		*BaseConfInfo.KeyFile = os.Getenv("KEY_FILE")
	}
	
	if os.Getenv("CA_FILE") != "" {
		*BaseConfInfo.CaFile = os.Getenv("CA_FILE")
	}
	
	if os.Getenv("MEDIA_TYPE") != "" {
		*BaseConfInfo.MediaType = os.Getenv("MEDIA_TYPE")
	}
	
	for _, userIdStr := range strings.Split(*allowedUserIds, ",") {
		if userIdStr == "" {
			continue
		}
		BaseConfInfo.AllowedTelegramUserIds[userIdStr] = true
	}
	
	for _, groupIdStr := range strings.Split(*allowedGroupIds, ",") {
		groupId, err := strconv.Atoi(groupIdStr)
		if err != nil {
			logger.Warn("AllowedTelegramGroupIds parse error", "groupId", groupIdStr)
			continue
		}
		BaseConfInfo.AllowedTelegramGroupIds[int64(groupId)] = true
	}
	
	for _, userIdStr := range strings.Split(*adminUserIds, ",") {
		if userIdStr == "" {
			continue
		}
		BaseConfInfo.AdminUserIds[userIdStr] = true
	}
	
	logger.Info("CONF", "TelegramBotToken", *BaseConfInfo.TelegramBotToken)
	logger.Info("CONF", "DiscordBotToken", *BaseConfInfo.DiscordBotToken)
	logger.Info("CONF", "DeepseekToken", *BaseConfInfo.DeepseekToken)
	logger.Info("CONF", "CustomUrl", *BaseConfInfo.CustomUrl)
	logger.Info("CONF", "Type", *BaseConfInfo.Type)
	logger.Info("CONF", "VolcAK", *BaseConfInfo.VolcAK)
	logger.Info("CONF", "VolcSK", *BaseConfInfo.VolcSK)
	logger.Info("CONF", "DBType", *BaseConfInfo.DBType)
	logger.Info("CONF", "DBConf", *BaseConfInfo.DBConf)
	logger.Info("CONF", "AllowedTelegramUserIds", *allowedUserIds)
	logger.Info("CONF", "AllowedTelegramGroupIds", *allowedGroupIds)
	logger.Info("CONF", "DeepseekProxy", *BaseConfInfo.DeepseekProxy)
	logger.Info("CONF", "TelegramProxy", *BaseConfInfo.TelegramProxy)
	logger.Info("CONF", "Lang", *BaseConfInfo.Lang)
	logger.Info("CONF", "TokenPerUser", *BaseConfInfo.TokenPerUser)
	logger.Info("CONF", "AdminUserIds", *adminUserIds)
	logger.Info("CONF", "NeedATBOt", *BaseConfInfo.NeedATBOt)
	logger.Info("CONF", "MaxUserChat", *BaseConfInfo.MaxUserChat)
	logger.Info("CONF", "VideoToken", *BaseConfInfo.VideoToken)
	logger.Info("CONF", "HTTPPort", *BaseConfInfo.HTTPPort)
	logger.Info("CONF", "OpenAIToken", *BaseConfInfo.OpenAIToken)
	logger.Info("CONF", "GeminiToken", *BaseConfInfo.GeminiToken)
	logger.Info("CONF", "OpenRouterToken", *BaseConfInfo.OpenRouterToken)
	logger.Info("CONF", "ErnieAK", *BaseConfInfo.ErnieAK)
	logger.Info("CONF", "ErnieSK", *BaseConfInfo.ErnieSK)
	logger.Info("CONF", "VolToken", *BaseConfInfo.VolToken)
	logger.Info("CONF", "CrtFile", *BaseConfInfo.CrtFile)
	logger.Info("CONF", "KeyFile", *BaseConfInfo.KeyFile)
	logger.Info("CONF", "CaFile", *BaseConfInfo.CaFile)
	logger.Info("CONF", "MediaType", *BaseConfInfo.MediaType)
	
	EnvAudioConf()
	EnvRagConf()
	EnvDeepseekConf()
	EnvPhotoConf()
	EnvToolsConf()
	EnvVideoConf()
	
	if *BaseConfInfo.TelegramBotToken == "" {
		panic("Bot token and llm token are required")
	}
	
}
