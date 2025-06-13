package conf

import (
	"os"
	"testing"
)

func TestInitConf_AllEnvVars(t *testing.T) {
	// 准备环境变量
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("CUSTOM_URL", "https://example.com")
	os.Setenv("TYPE", "pro")
	os.Setenv("VOLC_AK", "volc-ak")
	os.Setenv("VOLC_SK", "volc-sk")
	os.Setenv("DB_TYPE", "mysql")
	os.Setenv("DB_CONF", "user:pass@tcp(127.0.0.1:3306)/dbname")
	os.Setenv("ALLOWED_TELEGRAM_USER_IDS", "1001,1002")
	os.Setenv("ALLOWED_TELEGRAM_GROUP_IDS", "-2001,-2002")
	os.Setenv("DEEPSEEK_PROXY", "http://proxy.deepseek")
	os.Setenv("TELEGRAM_PROXY", "http://proxy.telegram")
	os.Setenv("LANG", "zh-CN")
	os.Setenv("TOKEN_PER_USER", "888")
	os.Setenv("ADMIN_USER_IDS", "9999,8888")
	os.Setenv("NEED_AT_BOT", "true")
	os.Setenv("MAX_USER_CHAT", "10")
	os.Setenv("VIDEO_TOKEN", "video_token_abc")
	os.Setenv("HTTP_PORT", "8888")
	os.Setenv("USE_TOOLS", "false")
	os.Setenv("OPENAI_TOKEN", "openai_test")
	os.Setenv("GEMINI_TOKEN", "gemini_test")
	os.Setenv("ERNIE_AK", "ernie-ak")
	os.Setenv("ERNIE_SK", "ernie-sk")

	os.Setenv("AUDIO_APP_ID", "test-audio-app-id")
	os.Setenv("AUDIO_TOKEN", "test-audio-token")
	os.Setenv("AUDIO_CLUSTER", "test-cluster")

	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("FREQUENCY_PENALTY", "0.5")
	os.Setenv("MAX_TOKENS", "2048")
	os.Setenv("PRESENCE_PENALTY", "1.0")
	os.Setenv("TEMPERATURE", "0.9")
	os.Setenv("TOP_P", "0.8")
	os.Setenv("STOP", "stop-sequence")
	os.Setenv("LOG_PROBS", "true")
	os.Setenv("TOP_LOG_PROBS", "5")

	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("REQ_KEY", "test-req-key")
	os.Setenv("MODEL_VERSION", "v2.1")
	os.Setenv("REQ_SCHEDULE_CONF", "scheduleA")
	os.Setenv("SEED", "1234")
	os.Setenv("SCALE", "2.5")
	os.Setenv("DDIM_Steps", "30")
	os.Setenv("WIDTH", "512")
	os.Setenv("Height", "768")
	os.Setenv("UsePreLLM", "true")
	os.Setenv("UseSr", "false")
	os.Setenv("ReturnUrl", "true")
	os.Setenv("AddLogo", "false")
	os.Setenv("Position", "bottom-right")
	os.Setenv("Language", "1")
	os.Setenv("Opacity", "0.75")
	os.Setenv("LogoTextContent", "Test Logo")

	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("EMBEDDING_TYPE", "openai")
	os.Setenv("KNOWLEDGE_PATH", "/data/knowledge")
	os.Setenv("VECTOR_DB_TYPE", "chroma")
	os.Setenv("CHROMA_URL", "http://localhost:8000")
	os.Setenv("SPACE", "test-space")
	os.Setenv("CHUNK_SIZE", "500")
	os.Setenv("CHUNK_OVERLAP", "50")

	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("AMAP_API_KEY", "amap-key")
	os.Setenv("GITHUB_ACCESS_TOKEN", "gh-token")
	os.Setenv("VM_URL", "http://vm-url")
	os.Setenv("VM_INSERT_URL", "http://vm-insert")
	os.Setenv("VM_SELECT_URL", "http://vm-select")
	os.Setenv("BINANCE_SWITCH", "1")
	os.Setenv("TIME_ZONE", "Asia/Shanghai")
	os.Setenv("PLAY_WRIGHT_SSE_SERVER", "http://sse")
	os.Setenv("PLAY_WRIGHT_SWITCH", "true")
	os.Setenv("FILECRAWL_API_KEY", "fc-api-key")
	os.Setenv("FILE_PATH", "/tmp/files")
	os.Setenv("GOOGLE_MAP_API_KEY", "gmap-key")
	os.Setenv("NOTION_AUTHORIZATION", "notion-auth")
	os.Setenv("NOTION_VERSION", "2022-06-28")
	os.Setenv("ALIYUN_ACCESS_KEY_ID", "aliyun-id")
	os.Setenv("ALIYUN_ACCESS_KEY_SECRET", "aliyun-secret")
	os.Setenv("AIRBNB_SWITCH", "yes")
	os.Setenv("BITCOIN_SWITCH", "true")
	os.Setenv("TWITTER_API_KEY", "twitter-key")
	os.Setenv("TWITTER_API_KEY_SECRET", "twitter-secret")
	os.Setenv("TWITTER_ACCESS_TOKEN", "twitter-token")
	os.Setenv("TWITTER_ACCESS_TOKEN_SECRET", "twitter-token-secret")
	os.Setenv("WHATSAPP_PATH", "/wa/path")
	os.Setenv("WHATSAPP_PYTHON_MAIN_FILE", "main.py")
	os.Setenv("BAIDUMAP_API_KEY", "baidu-key")

	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("VIDEO_MODEL", "model-v1")
	os.Setenv("RADIO", "radio-123")
	os.Setenv("DURATION", "120")
	os.Setenv("FPS", "30")
	os.Setenv("RESOLUTION", "1920x1080")
	os.Setenv("WATERMARK", "true")

	// 调用初始化函数
	InitConf()

	// 断言检查
	assertEqual(t, *BotToken, "test_bot_token", "BotToken")
	assertEqual(t, *DeepseekToken, "test_deepseek_token", "DeepseekToken")
	assertEqual(t, *CustomUrl, "https://example.com", "CustomUrl")
	assertEqual(t, *Type, "pro", "Type")
	assertEqual(t, *VolcAK, "volc-ak", "VolcAK")
	assertEqual(t, *VolcSK, "volc-sk", "VolcSK")
	assertEqual(t, *DBType, "mysql", "DBType")
	assertEqual(t, *DBConf, "user:pass@tcp(127.0.0.1:3306)/dbname", "DBConf")
	assertEqual(t, *DeepseekProxy, "http://proxy.deepseek", "DeepseekProxy")
	assertEqual(t, *TelegramProxy, "http://proxy.telegram", "TelegramProxy")
	assertEqual(t, *Lang, "zh-CN", "Lang")
	assertInt(t, *TokenPerUser, 888, "TokenPerUser")
	assertBool(t, *NeedATBOt, true, "NeedATBOt")
	assertInt(t, *MaxUserChat, 10, "MaxUserChat")
	assertEqual(t, *VideoToken, "video_token_abc", "VideoToken")
	assertInt(t, *HTTPPort, 8888, "HTTPPort")
	assertBool(t, *UseTools, false, "UseTools")
	assertEqual(t, *OpenAIToken, "openai_test", "OpenAIToken")
	assertEqual(t, *GeminiToken, "gemini_test", "GeminiToken")
	assertEqual(t, *ErnieAK, "ernie-ak", "ErnieAK")
	assertEqual(t, *ErnieSK, "ernie-sk", "ErnieSK")

	assertEqual(t, *AudioAppID, "test-audio-app-id", "AudioAppID")
	assertEqual(t, *AudioToken, "test-audio-token", "AudioToken")
	assertEqual(t, *AudioCluster, "test-cluster", "AudioCluster")

	assertFloatEqual(t, *FrequencyPenalty, 0.5, "FrequencyPenalty")
	assertInt(t, *MaxTokens, 2048, "MaxTokens")
	assertFloatEqual(t, *PresencePenalty, 1.0, "PresencePenalty")
	assertFloatEqual(t, *Temperature, 0.9, "Temperature")
	assertFloatEqual(t, *TopP, 0.8, "TopP")
	assertBool(t, *LogProbs, true, "LogProbs")
	assertInt(t, *TopLogProbs, 5, "TopLogProbs")

	assertEqual(t, *ReqKey, "test-req-key", "ReqKey")
	assertEqual(t, *ModelVersion, "v2.1", "ModelVersion")
	assertEqual(t, *ReqScheduleConf, "scheduleA", "ReqScheduleConf")
	assertInt(t, *Seed, 1234, "Seed")
	assertFloatEqual(t, *Scale, 2.5, "Scale")
	assertInt(t, *DDIMSteps, 30, "DDIMSteps")
	assertInt(t, *Width, 512, "Width")
	assertInt(t, *Height, 768, "Height")
	assertBool(t, *UsePreLLM, true, "UsePreLLM")
	assertBool(t, *UseSr, false, "UseSr")
	assertBool(t, *ReturnUrl, true, "ReturnUrl")
	assertBool(t, *AddLogo, false, "AddLogo")
	assertEqual(t, *Position, "bottom-right", "Position")
	assertInt(t, *Language, 1, "Language")
	assertFloatEqual(t, *Opacity, 0.75, "Opacity")
	assertEqual(t, *LogoTextContent, "Test Logo", "LogoTextContent")

	assertEqual(t, *EmbeddingType, "openai", "EmbeddingType")
	assertEqual(t, *KnowledgePath, "/data/knowledge", "KnowledgePath")
	assertEqual(t, *VectorDBType, "chroma", "VectorDBType")
	assertEqual(t, *ChromaURL, "http://localhost:8000", "ChromaURL")
	assertEqual(t, *Space, "test-space", "ChromaSpace")
	assertInt(t, *ChunkSize, 500, "ChunkSize")
	assertInt(t, *ChunkOverlap, 50, "ChunkOverlap")

	assertEqual(t, *AmapApiKey, "amap-key", "AMAP_API_KEY")
	assertEqual(t, *GithubAccessToken, "gh-token", "GITHUB_ACCESS_TOKEN")
	assertEqual(t, *VMUrl, "http://vm-url", "VM_URL")
	assertEqual(t, *VMInsertUrl, "http://vm-insert", "VM_INSERT_URL")
	assertEqual(t, *VMSelectUrl, "http://vm-select", "VM_SELECT_URL")
	assertBool(t, *BinanceSwitch, true, "BINANCE_SWITCH")
	assertEqual(t, *TimeZone, "Asia/Shanghai", "TIME_ZONE")
	assertEqual(t, *PlayWrightSSEServer, "http://sse", "PLAY_WRIGHT_SSE_SERVER")
	assertBool(t, *PlayWrightSwitch, true, "PLAY_WRIGHT_SWITCH")
	assertEqual(t, *FilecrawlApiKey, "fc-api-key", "FILECRAWL_API_KEY")
	assertEqual(t, *FilePath, "/tmp/files", "FILE_PATH")
	assertEqual(t, *GoogleMapApiKey, "gmap-key", "GOOGLE_MAP_API_KEY")
	assertEqual(t, *NotionAuthorization, "notion-auth", "NOTION_AUTHORIZATION")
	assertEqual(t, *NotionVersion, "2022-06-28", "NOTION_VERSION")
	assertEqual(t, *AliyunAccessKeyID, "aliyun-id", "ALIYUN_ACCESS_KEY_ID")
	assertEqual(t, *AliyunAccessKeySecret, "aliyun-secret", "ALIYUN_ACCESS_KEY_SECRET")
	assertBool(t, *AirBnbSwitch, true, "AIRBNB_SWITCH")
	assertBool(t, *BitCoinSwitch, true, "BITCOIN_SWITCH")
	assertEqual(t, *TwitterApiKey, "twitter-key", "TWITTER_API_KEY")
	assertEqual(t, *TwitterApiSecretKey, "twitter-secret", "TWITTER_API_KEY_SECRET")
	assertEqual(t, *TwitterAccessToken, "twitter-token", "TWITTER_ACCESS_TOKEN")
	assertEqual(t, *TwitterAccessTokenSecret, "twitter-token-secret", "TWITTER_ACCESS_TOKEN_SECRET")
	assertEqual(t, *WhatsappPath, "/wa/path", "WHATSAPP_PATH")
	assertEqual(t, *WhatsappPythonMainFile, "main.py", "WHATSAPP_PYTHON_MAIN_FILE")
	assertEqual(t, *BaidumapApiKey, "baidu-key", "BAIDUMAP_API_KEY")

	assertEqual(t, *VideoModel, "model-v1", "VIDEO_MODEL")
	assertEqual(t, *Radio, "radio-123", "RADIO")
	assertInt(t, *Duration, 120, "DURATION")
	assertInt(t, *FPS, 30, "FPS")
	assertEqual(t, *Resolution, "1920x1080", "RESOLUTION")
	assertBool(t, *Watermark, true, "WATERMARK")

	os.Clearenv()
}

// 辅助函数
func assertEqual(t *testing.T, got, expected, field string) {
	if got != expected {
		t.Errorf("%s expected '%s', got '%s'", field, expected, got)
	}
}

func assertInt(t *testing.T, got int, expected int, field string) {
	if got != expected {
		t.Errorf("%s expected %d, got %d", field, expected, got)
	}
}

func assertBool(t *testing.T, got bool, expected bool, field string) {
	if got != expected {
		t.Errorf("%s expected %v, got %v", field, expected, got)
	}
}

func assertFloatEqual(t *testing.T, got, expected float64, field string) {
	if got != expected {
		t.Errorf("%s expected %.2f, got %.2f", field, expected, got)
	}
}
