package conf

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	
	"github.com/yincongcyincong/MuseBot/logger"
)

type BaseConf struct {
	StartTime int64 `json:"-"`
	
	TelegramBotToken        *string `json:"telegram_bot_token"`
	DiscordBotToken         *string `json:"discord_bot_token"`
	SlackBotToken           *string `json:"slack_bot_token"`
	SlackAppToken           *string `json:"slack_app_token"`
	LarkAPPID               *string `json:"lark_app_id"`
	LarkAppSecret           *string `json:"lark_app_secret"`
	DingClientId            *string `json:"ding_client_id"`
	DingClientSecret        *string `json:"ding_client_secret"`
	DingTemplateId          *string `json:"ding_template_id"`
	ComWechatToken          *string `json:"com_wechat_token"`
	ComWechatEncodingAESKey *string `json:"com_wechat_encoding_aes_key"`
	ComWechatCorpID         *string `json:"com_wechat_corp_id"`
	ComWechatSecret         *string `json:"com_wechat_secret"`
	ComWechatAgentID        *string `json:"com_wechat_agent_id"`
	WechatAppID             *string `json:"wechat_app_id"`
	WechatAppSecret         *string `json:"wechat_app_secret"`
	WechatToken             *string `json:"wechat_token"`
	WechatEncodingAESKey    *string `json:"wechat_encoding_aes_key"`
	WechatActive            *bool   `json:"wechat_active"`
	QQAppID                 *string `json:"qq_app_id"`
	QQAppSecret             *string `json:"qq_app_secret"`
	QQNapCatReceiveToken    *string `json:"qq_nap_cat_check_token"`
	QQNapCatSendToken       *string `json:"qq_nap_cat_send_token"`
	QQNapCatHttpServer      *string `json:"qq_nap_cat_http_server"`
	
	DeepseekToken     *string `json:"deepseek_token"`
	OpenAIToken       *string `json:"openai_token"`
	GeminiToken       *string `json:"gemini_token"`
	OpenRouterToken   *string `json:"open_router_token"`
	AI302Token        *string `json:"ai_302_token"`
	VolToken          *string `json:"vol_token"`
	AliyunToken       *string `json:"aliyun_token"`
	ChatAnyWhereToken *string `json:"chat_any_where_token"`
	ErnieAK           *string `json:"ernie_ak"`
	ErnieSK           *string `json:"ernie_sk"`
	
	BotName      *string `json:"bot_name"`
	Type         *string `json:"type"`
	MediaType    *string `json:"media_type"`
	CustomUrl    *string `json:"custom_url"`
	CustomPath   *string `json:"custom_path"`
	VolcAK       *string `json:"volc_ak"`
	VolcSK       *string `json:"volc_sk"`
	DBType       *string `json:"db_type"`
	DBConf       *string `json:"db_conf"`
	LLMProxy     *string `json:"llm_proxy"`
	RobotProxy   *string `json:"robot_proxy"`
	Lang         *string `json:"lang"`
	TokenPerUser *int    `json:"token_per_user"`
	MaxUserChat  *int    `json:"max_user_chat"`
	HTTPHost     *string `json:"http_host"`
	UseTools     *bool   `json:"use_tools"`
	MaxQAPair    *int    `json:"max_qa_pari"`
	Character    *string `json:"character"`
	
	CrtFile *string `json:"crt_file"`
	KeyFile *string `json:"key_file"`
	CaFile  *string `json:"ca_file"`
	
	AllowedUserIds  map[string]bool `json:"allowed_user_ids"`
	AllowedGroupIds map[string]bool `json:"allowed_group_ids"`
	AdminUserIds    map[string]bool `json:"admin_user_ids"`
}

var (
	BaseConfInfo = new(BaseConf)
)

func InitConf() {
	BaseConfInfo.StartTime = time.Now().Unix()
	BaseConfInfo.TelegramBotToken = flag.String("telegram_bot_token", "", "Telegram bot tokens")
	BaseConfInfo.DiscordBotToken = flag.String("discord_bot_token", "", "Discord bot tokens")
	BaseConfInfo.SlackBotToken = flag.String("slack_bot_token", "", "Slack bot tokens")
	BaseConfInfo.SlackAppToken = flag.String("slack_app_token", "", "Slack app tokens")
	BaseConfInfo.LarkAPPID = flag.String("lark_app_id", "", "Lark app id")
	BaseConfInfo.LarkAppSecret = flag.String("lark_app_secret", "", "Lark app secret")
	BaseConfInfo.DingClientId = flag.String("ding_client_id", "", "Dingding client id")
	BaseConfInfo.DingClientSecret = flag.String("ding_client_secret", "", "Dingding app secret")
	BaseConfInfo.DingTemplateId = flag.String("ding_template_id", "", "Dingding template id")
	BaseConfInfo.ComWechatToken = flag.String("com_wechat_token", "", "ComWechat token")
	BaseConfInfo.ComWechatEncodingAESKey = flag.String("com_wechat_encoding_aes_key", "", "ComWechat encoding aes key")
	BaseConfInfo.ComWechatCorpID = flag.String("com_wechat_corp_id", "", "ComWechat corp id")
	BaseConfInfo.ComWechatSecret = flag.String("com_wechat_secret", "", "ComWechat secret")
	BaseConfInfo.ComWechatAgentID = flag.String("com_wechat_agent_id", "", "ComWechat agent id")
	BaseConfInfo.WechatAppID = flag.String("wechat_app_id", "", "Wechat app id")
	BaseConfInfo.WechatAppSecret = flag.String("wechat_app_secret", "", "Wechat app secret")
	BaseConfInfo.WechatEncodingAESKey = flag.String("wechat_encoding_aes_key", "", "Wechat encoding aes key")
	BaseConfInfo.WechatToken = flag.String("wechat_token", "", "Wechat token")
	BaseConfInfo.WechatActive = flag.Bool("wechat_active", false, "Wechat active")
	BaseConfInfo.QQAppID = flag.String("qq_app_id", "", "QQ app id")
	BaseConfInfo.QQAppSecret = flag.String("qq_app_secret", "", "QQ app secret")
	BaseConfInfo.QQNapCatReceiveToken = flag.String("qq_napcat_receive_token", "MuseBot", "napcat receive token")
	BaseConfInfo.QQNapCatSendToken = flag.String("qq_napcat_send_token", "MuseBot", "napcat send token")
	BaseConfInfo.QQNapCatHttpServer = flag.String("qq_napcat_http_server", "http://127.0.0.1:3000", "napcat http server")
	
	BaseConfInfo.DeepseekToken = flag.String("deepseek_token", "", "deepseek auth token")
	BaseConfInfo.OpenAIToken = flag.String("openai_token", "", "openai auth token")
	BaseConfInfo.GeminiToken = flag.String("gemini_token", "", "gemini auth token")
	BaseConfInfo.OpenRouterToken = flag.String("open_router_token", "", "openrouter auth token")
	BaseConfInfo.AI302Token = flag.String("ai_302_token", "", "302.ai token")
	BaseConfInfo.VolToken = flag.String("vol_token", "", "vol auth token")
	BaseConfInfo.AliyunToken = flag.String("aliyun_token", "", "aliyun auth token")
	BaseConfInfo.ErnieAK = flag.String("ernie_ak", "", "ernie ak")
	BaseConfInfo.ErnieSK = flag.String("ernie_sk", "", "ernie sk")
	BaseConfInfo.VolcAK = flag.String("volc_ak", "", "volc ak")
	BaseConfInfo.VolcSK = flag.String("volc_sk", "", "volc sk")
	BaseConfInfo.ChatAnyWhereToken = flag.String("chat_any_where_token", "", "chatAnyWhere Token")
	
	BaseConfInfo.BotName = flag.String("bot_name", "MuseBot", "bot name")
	BaseConfInfo.CustomUrl = flag.String("custom_url", "", "custom url")
	BaseConfInfo.CustomPath = flag.String("custom_path", "", "custom path")
	BaseConfInfo.Type = flag.String("type", "", "llm type: deepseek gemini openai openrouter vol chatanywhere")
	BaseConfInfo.MediaType = flag.String("media_type", "", "media type: vol gemini openai aliyun 302-ai openrouter")
	BaseConfInfo.DBType = flag.String("db_type", "sqlite3", "db type")
	BaseConfInfo.DBConf = flag.String("db_conf", GetAbsPath("data/muse_bot.db"), "db conf")
	BaseConfInfo.LLMProxy = flag.String("llm_proxy", "", "llm proxy: http://127.0.0.1:7890")
	BaseConfInfo.RobotProxy = flag.String("robot_proxy", "", "robot proxy: http://127.0.0.1:7890")
	BaseConfInfo.Lang = flag.String("lang", "en", "lang")
	BaseConfInfo.TokenPerUser = flag.Int("token_per_user", 10000, "token per user")
	BaseConfInfo.MaxUserChat = flag.Int("max_user_chat", 2, "max chat per user")
	BaseConfInfo.HTTPHost = flag.String("http_host", ":36060", "http server port")
	BaseConfInfo.UseTools = flag.Bool("use_tools", false, "use function tools")
	BaseConfInfo.MaxQAPair = flag.Int("max_qa_pari", 100, "max qa pair")
	BaseConfInfo.Character = flag.String("character", "", "ai's character")
	
	BaseConfInfo.CrtFile = flag.String("crt_file", "", "public key file")
	BaseConfInfo.KeyFile = flag.String("key_file", "", "secret key file")
	BaseConfInfo.CaFile = flag.String("ca_file", "", "ca file")
	
	adminUserIds := flag.String("admin_user_ids", "", "admin user ids")
	allowedUserIds := flag.String("allowed_user_ids", "", "allowed user ids")
	allowedGroupIds := flag.String("allowed_group_ids", "", "allowed group ids")
	
	BaseConfInfo.AllowedUserIds = make(map[string]bool)
	BaseConfInfo.AllowedGroupIds = make(map[string]bool)
	BaseConfInfo.AdminUserIds = make(map[string]bool)
	
	InitDeepseekConf()
	InitPhotoConf()
	InitVideoConf()
	InitAudioConf()
	InitToolsConf()
	InitRagConf()
	InitRegisterConf()
	
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.Parse()
	
	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" {
		*BaseConfInfo.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	
	if os.Getenv("CHAT_ANY_WHERE_TOKEN") != "" {
		*BaseConfInfo.ChatAnyWhereToken = os.Getenv("CHAT_ANY_WHERE_TOKEN")
	}
	
	if os.Getenv("DISCORD_BOT_TOKEN") != "" {
		*BaseConfInfo.DiscordBotToken = os.Getenv("DISCORD_BOT_TOKEN")
	}
	
	if os.Getenv("SLACK_BOT_TOKEN") != "" {
		*BaseConfInfo.SlackBotToken = os.Getenv("SLACK_BOT_TOKEN")
	}
	
	if os.Getenv("SLACK_APP_TOKEN") != "" {
		*BaseConfInfo.SlackAppToken = os.Getenv("SLACK_APP_TOKEN")
	}
	
	if os.Getenv("LARK_APP_ID") != "" {
		*BaseConfInfo.LarkAPPID = os.Getenv("LARK_APP_ID")
	}
	
	if os.Getenv("LARK_APP_SECRET") != "" {
		*BaseConfInfo.LarkAppSecret = os.Getenv("LARK_APP_SECRET")
	}
	
	if os.Getenv("DING_CLIENT_ID") != "" {
		*BaseConfInfo.DingClientId = os.Getenv("DING_CLIENT_ID")
	}
	
	if os.Getenv("DING_CLIENT_SECRET") != "" {
		*BaseConfInfo.DingClientSecret = os.Getenv("DING_CLIENT_SECRET")
	}
	
	if os.Getenv("DING_TEMPLATE_ID") != "" {
		*BaseConfInfo.DingTemplateId = os.Getenv("DING_TEMPLATE_ID")
	}
	
	if os.Getenv("COM_WECHAT_TOKEN") != "" {
		*BaseConfInfo.ComWechatToken = os.Getenv("COM_WECHAT_TOKEN")
	}
	
	if os.Getenv("WECHAT_TOKEN") != "" {
		*BaseConfInfo.WechatToken = os.Getenv("WECHAT_TOKEN")
	}
	
	if os.Getenv("WECHAT_APP_ID") != "" {
		*BaseConfInfo.WechatAppID = os.Getenv("WECHAT_APP_ID")
	}
	
	if os.Getenv("WECHAT_APP_SECRET") != "" {
		*BaseConfInfo.WechatAppSecret = os.Getenv("WECHAT_APP_SECRET")
	}
	
	if os.Getenv("WECHAT_ENCODING_AES_KEY") != "" {
		*BaseConfInfo.WechatEncodingAESKey = os.Getenv("WECHAT_ENCODING_AES_KEY")
	}
	
	if os.Getenv("WECHAT_ACTIVE") != "" {
		*BaseConfInfo.WechatActive = os.Getenv("WECHAT_ACTIVE") == "true"
	}
	
	if os.Getenv("COM_WECHAT_ENCODING_AES_KEY") != "" {
		*BaseConfInfo.ComWechatEncodingAESKey = os.Getenv("COM_WECHAT_ENCODING_AES_KEY")
	}
	
	if os.Getenv("COM_WECHAT_CORP_ID") != "" {
		*BaseConfInfo.ComWechatCorpID = os.Getenv("COM_WECHAT_CORP_ID")
	}
	
	if os.Getenv("COM_WECHAT_SECRET") != "" {
		*BaseConfInfo.ComWechatSecret = os.Getenv("COM_WECHAT_SECRET")
	}
	
	if os.Getenv("COM_WECHAT_AGENT_ID") != "" {
		*BaseConfInfo.ComWechatAgentID = os.Getenv("COM_WECHAT_AGENT_ID")
	}
	
	if os.Getenv("QQ_APP_ID") != "" {
		*BaseConfInfo.QQAppID = os.Getenv("QQ_APP_ID")
	}
	
	if os.Getenv("QQ_APP_SECRET") != "" {
		*BaseConfInfo.QQAppSecret = os.Getenv("QQ_APP_SECRET")
	}
	
	if os.Getenv("QQ_NAPCAT_SEND_TOKEN") != "" {
		*BaseConfInfo.QQNapCatSendToken = os.Getenv("QQ_NAPCAT_SEND_TOKEN")
	}
	
	if os.Getenv("QQ_NAPCAT_RECEIVE_TOKEN") != "" {
		*BaseConfInfo.QQNapCatReceiveToken = os.Getenv("QQ_NAPCAT_RECEIVE_TOKEN")
	}
	
	if os.Getenv("QQ_NAPCAT_HTTP_SERVER") != "" {
		*BaseConfInfo.QQNapCatHttpServer = os.Getenv("QQ_NAPCAT_HTTP_SERVER")
	}
	
	if os.Getenv("DEEPSEEK_TOKEN") != "" {
		*BaseConfInfo.DeepseekToken = os.Getenv("DEEPSEEK_TOKEN")
	}
	
	if os.Getenv("CUSTOM_URL") != "" {
		*BaseConfInfo.CustomUrl = os.Getenv("CUSTOM_URL")
	}
	
	if os.Getenv("BOT_NAME") != "" {
		*BaseConfInfo.BotName = os.Getenv("BOT_NAME")
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
	
	if os.Getenv("ALLOWED_USER_IDS") != "" {
		*allowedUserIds = os.Getenv("ALLOWED_USER_IDS")
	}
	
	if os.Getenv("ALLOWED_GROUP_IDS") != "" {
		*allowedGroupIds = os.Getenv("ALLOWED_GROUP_IDS")
	}
	
	if os.Getenv("LLM_PROXY") != "" {
		*BaseConfInfo.LLMProxy = os.Getenv("LLM_PROXY")
	}
	
	if os.Getenv("ROBOT_PROXY") != "" {
		*BaseConfInfo.RobotProxy = os.Getenv("ROBOT_PROXY")
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
	
	if os.Getenv("MAX_USER_CHAT") != "" {
		*BaseConfInfo.MaxUserChat, _ = strconv.Atoi(os.Getenv("MAX_USER_CHAT"))
	}
	
	if os.Getenv("HTTP_HOST") != "" {
		*BaseConfInfo.HTTPHost = os.Getenv("HTTP_HOST")
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
	
	if os.Getenv("ALIYUN_TOKEN") != "" {
		*BaseConfInfo.AliyunToken = os.Getenv("ALIYUN_TOKEN")
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
	
	if os.Getenv("AI_302_TOKEN") != "" {
		*BaseConfInfo.AI302Token = os.Getenv("AI_302_TOKEN")
	}
	
	if os.Getenv("MAX_QA_PAIR") != "" {
		*BaseConfInfo.MaxQAPair, _ = strconv.Atoi(os.Getenv("MAX_QA_PAIR"))
	}
	
	if os.Getenv("CHARACTER") != "" {
		*BaseConfInfo.Character = os.Getenv("CHARACTER")
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
		BaseConfInfo.AllowedUserIds[userIdStr] = true
	}
	
	for _, groupIdStr := range strings.Split(*allowedGroupIds, ",") {
		BaseConfInfo.AllowedGroupIds[groupIdStr] = true
	}
	
	for _, userIdStr := range strings.Split(*adminUserIds, ",") {
		if userIdStr == "" {
			continue
		}
		BaseConfInfo.AdminUserIds[userIdStr] = true
	}
	
	logger.Info("CONF", "TelegramBotToken", *BaseConfInfo.TelegramBotToken)
	logger.Info("CONF", "DiscordBotToken", *BaseConfInfo.DiscordBotToken)
	logger.Info("CONF", "SlackBotToken", *BaseConfInfo.SlackBotToken)
	logger.Info("CONF", "SlackAppToken", *BaseConfInfo.SlackAppToken)
	logger.Info("CONF", "LarkAPPID", *BaseConfInfo.LarkAPPID)
	logger.Info("CONF", "LarkAppSecret", *BaseConfInfo.LarkAppSecret)
	logger.Info("CONF", "DingClientId", *BaseConfInfo.DingClientId)
	logger.Info("CONF", "DingClientSecret", *BaseConfInfo.DingClientSecret)
	logger.Info("CONF", "DingTemplateId", *BaseConfInfo.DingTemplateId)
	logger.Info("CONF", "ComWechatToken", *BaseConfInfo.ComWechatToken)
	logger.Info("CONF", "ComWechatEncodingAESKey", *BaseConfInfo.ComWechatEncodingAESKey)
	logger.Info("CONF", "ComWechatCorpID", *BaseConfInfo.ComWechatCorpID)
	logger.Info("CONF", "ComWechatSecret", *BaseConfInfo.ComWechatSecret)
	logger.Info("CONF", "ComWechatAgentID", *BaseConfInfo.ComWechatAgentID)
	logger.Info("CONF", "WechatToken", *BaseConfInfo.WechatToken)
	logger.Info("CONF", "WechatAppSecret", *BaseConfInfo.WechatAppSecret)
	logger.Info("CONF", "WechatAppID", *BaseConfInfo.WechatAppID)
	logger.Info("CONF", "WechatActive", *BaseConfInfo.WechatActive)
	logger.Info("CONF", "WechatEncodingAESKey", *BaseConfInfo.WechatEncodingAESKey)
	logger.Info("CONF", "QQAppID", *BaseConfInfo.QQAppID)
	logger.Info("CONF", "QQAppSecret", *BaseConfInfo.QQAppSecret)
	logger.Info("CONF", "QQNapCatHttpServer", *BaseConfInfo.QQNapCatHttpServer)
	logger.Info("CONF", "QQNapCatReceiveToken", *BaseConfInfo.QQNapCatReceiveToken)
	logger.Info("CONF", "QQNapCatSendToken", *BaseConfInfo.QQNapCatSendToken)
	logger.Info("CONF", "DeepseekToken", *BaseConfInfo.DeepseekToken)
	logger.Info("CONF", "CustomUrl", *BaseConfInfo.CustomUrl)
	logger.Info("CONF", "Type", *BaseConfInfo.Type)
	logger.Info("CONF", "VolcAK", *BaseConfInfo.VolcAK)
	logger.Info("CONF", "VolcSK", *BaseConfInfo.VolcSK)
	logger.Info("CONF", "AliyunToken", *BaseConfInfo.AliyunToken)
	logger.Info("CONF", "DBType", *BaseConfInfo.DBType)
	logger.Info("CONF", "DBConf", *BaseConfInfo.DBConf)
	logger.Info("CONF", "AllowedUserIds", *allowedUserIds)
	logger.Info("CONF", "AllowedGroupIds", *allowedGroupIds)
	logger.Info("CONF", "LLMProxy", *BaseConfInfo.LLMProxy)
	logger.Info("CONF", "RobotProxy", *BaseConfInfo.RobotProxy)
	logger.Info("CONF", "Lang", *BaseConfInfo.Lang)
	logger.Info("CONF", "TokenPerUser", *BaseConfInfo.TokenPerUser)
	logger.Info("CONF", "AdminUserIds", *adminUserIds)
	logger.Info("CONF", "MaxUserChat", *BaseConfInfo.MaxUserChat)
	logger.Info("CONF", "HTTPHost", *BaseConfInfo.HTTPHost)
	logger.Info("CONF", "OpenAIToken", *BaseConfInfo.OpenAIToken)
	logger.Info("CONF", "GeminiToken", *BaseConfInfo.GeminiToken)
	logger.Info("CONF", "OpenRouterToken", *BaseConfInfo.OpenRouterToken)
	logger.Info("CONF", "AI302Token", *BaseConfInfo.AI302Token)
	logger.Info("CONF", "ErnieAK", *BaseConfInfo.ErnieAK)
	logger.Info("CONF", "ErnieSK", *BaseConfInfo.ErnieSK)
	logger.Info("CONF", "VolToken", *BaseConfInfo.VolToken)
	logger.Info("CONF", "CrtFile", *BaseConfInfo.CrtFile)
	logger.Info("CONF", "KeyFile", *BaseConfInfo.KeyFile)
	logger.Info("CONF", "CaFile", *BaseConfInfo.CaFile)
	logger.Info("CONF", "MediaType", *BaseConfInfo.MediaType)
	logger.Info("CONF", "BotName", *BaseConfInfo.BotName)
	logger.Info("CONF", "MaxQAPair", *BaseConfInfo.MaxQAPair)
	
	EnvAudioConf()
	EnvRagConf()
	EnvDeepseekConf()
	EnvPhotoConf()
	EnvToolsConf()
	EnvVideoConf()
	EnvRegisterConf()
	
}

func GetAbsPath(relPath string) string {
	exe, err := os.Executable()
	if err != nil {
		logger.Error("Failed to get executable path", "err", err)
		return ""
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, relPath)
}
