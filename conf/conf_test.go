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
	os.Setenv("ALLOWED_USER_IDS", "1001,1002")
	os.Setenv("ALLOWED_GROUP_IDS", "-2001,-2002")
	os.Setenv("LLM_PROXY", "http://proxy.deepseek")
	os.Setenv("ROBOT_PROXY", "http://proxy.telegram")
	os.Setenv("LANG", "zh-CN")
	os.Setenv("TOKEN_PER_USER", "888")
	os.Setenv("ADMIN_USER_IDS", "9999,8888")
	os.Setenv("NEED_AT_BOT", "true")
	os.Setenv("MAX_USER_CHAT", "10")
	os.Setenv("VIDEO_TOKEN", "video_token_abc")
	os.Setenv("HTTP_HOST", "8888")
	os.Setenv("USE_TOOLS", "false")
	os.Setenv("OPENAI_TOKEN", "openai_test")
	os.Setenv("GEMINI_TOKEN", "gemini_test")
	os.Setenv("ERNIE_AK", "ernie-ak")
	os.Setenv("ERNIE_SK", "ernie-sk")

	os.Setenv("VOL_AUDIO_APP_ID", "test-audio-app-id")
	os.Setenv("VOL_AUDIO_TOKEN", "test-audio-token")
	os.Setenv("VOL_AUDIO_REC_CLUSTER", "test-cluster")

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
	os.Setenv("HEIGHT", "768")
	os.Setenv("USE_PRE_LLM", "true")
	os.Setenv("USE_SR", "false")
	os.Setenv("RETURN_URL", "true")
	os.Setenv("ADD_LOGO", "false")
	os.Setenv("POSITION", "bottom-right")
	os.Setenv("PHOTO_LANGUAGE", "1")
	os.Setenv("OPACITY", "0.75")
	os.Setenv("LOGO_TEXT_CONTENT", "Test Logo")

	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("EMBEDDING_TYPE", "openai")
	os.Setenv("KNOWLEDGE_PATH", "/data/knowledge")
	os.Setenv("VECTOR_DB_TYPE", "milvus")
	os.Setenv("CHROMA_URL", "http://localhost:8000")
	os.Setenv("SPACE", "test-space")
	os.Setenv("CHUNK_SIZE", "500")
	os.Setenv("CHUNK_OVERLAP", "50")

	os.Setenv("MCP_CONF_PATH", "./conf/mcp/mcp.json")

	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("VOL_VIDEO_MODEL", "model-v1")
	os.Setenv("RADIO", "radio-123")
	os.Setenv("DURATION", "120")
	os.Setenv("FPS", "30")
	os.Setenv("RESOLUTION", "1920x1080")
	os.Setenv("WATERMARK", "true")

	// 调用初始化函数
	InitConf()

	// 断言检查
	assertEqual(t, *BaseConfInfo.TelegramBotToken, "test_bot_token", "BotToken")
	assertEqual(t, *BaseConfInfo.DeepseekToken, "test_deepseek_token", "DeepseekToken")
	assertEqual(t, *BaseConfInfo.CustomUrl, "https://example.com", "CustomUrl")
	assertEqual(t, *BaseConfInfo.Type, "pro", "Type")
	assertEqual(t, *BaseConfInfo.VolcAK, "volc-ak", "VolcAK")
	assertEqual(t, *BaseConfInfo.VolcSK, "volc-sk", "VolcSK")
	assertEqual(t, *BaseConfInfo.DBType, "mysql", "DBType")
	assertEqual(t, *BaseConfInfo.DBConf, "user:pass@tcp(127.0.0.1:3306)/dbname", "DBConf")
	assertEqual(t, *BaseConfInfo.LLMProxy, "http://proxy.deepseek", "LLMProxy")
	assertEqual(t, *BaseConfInfo.RobotProxy, "http://proxy.telegram", "RobotProxy")
	assertEqual(t, *BaseConfInfo.Lang, "zh-CN", "Lang")
	assertInt(t, *BaseConfInfo.TokenPerUser, 888, "TokenPerUser")
	assertInt(t, *BaseConfInfo.MaxUserChat, 10, "MaxUserChat")
	assertEqual(t, *BaseConfInfo.HTTPHost, "8888", "HTTPPort")
	assertBool(t, *BaseConfInfo.UseTools, false, "UseTools")
	assertEqual(t, *BaseConfInfo.OpenAIToken, "openai_test", "OpenAIToken")
	assertEqual(t, *BaseConfInfo.GeminiToken, "gemini_test", "GeminiToken")
	assertEqual(t, *BaseConfInfo.ErnieAK, "ernie-ak", "ErnieAK")
	assertEqual(t, *BaseConfInfo.ErnieSK, "ernie-sk", "ErnieSK")

	assertEqual(t, *AudioConfInfo.VolAudioAppID, "test-audio-app-id", "AudioAppID")
	assertEqual(t, *AudioConfInfo.VolAudioToken, "test-audio-token", "AudioToken")
	assertEqual(t, *AudioConfInfo.VolAudioRecCluster, "test-cluster", "AudioCluster")

	assertFloatEqual(t, *LLMConfInfo.FrequencyPenalty, 0.5, "FrequencyPenalty")
	assertInt(t, *LLMConfInfo.MaxTokens, 2048, "MaxTokens")
	assertFloatEqual(t, *LLMConfInfo.PresencePenalty, 1.0, "PresencePenalty")
	assertFloatEqual(t, *LLMConfInfo.Temperature, 0.9, "Temperature")
	assertFloatEqual(t, *LLMConfInfo.TopP, 0.8, "TopP")
	assertBool(t, *LLMConfInfo.LogProbs, true, "LogProbs")
	assertInt(t, *LLMConfInfo.TopLogProbs, 5, "TopLogProbs")

	assertEqual(t, *PhotoConfInfo.ReqKey, "test-req-key", "ReqKey")
	assertEqual(t, *PhotoConfInfo.ModelVersion, "v2.1", "ModelVersion")
	assertEqual(t, *PhotoConfInfo.ReqScheduleConf, "scheduleA", "ReqScheduleConf")
	assertInt(t, *PhotoConfInfo.Seed, 1234, "Seed")
	assertFloatEqual(t, *PhotoConfInfo.Scale, 2.5, "Scale")
	assertInt(t, *PhotoConfInfo.DDIMSteps, 30, "DDIMSteps")
	assertInt(t, *PhotoConfInfo.Width, 512, "Width")
	assertInt(t, *PhotoConfInfo.Height, 768, "Height")
	assertBool(t, *PhotoConfInfo.UsePreLLM, true, "UsePreLLM")
	assertBool(t, *PhotoConfInfo.UseSr, false, "UseSr")
	assertBool(t, *PhotoConfInfo.ReturnUrl, true, "ReturnUrl")
	assertBool(t, *PhotoConfInfo.AddLogo, false, "AddLogo")
	assertEqual(t, *PhotoConfInfo.Position, "bottom-right", "Position")
	assertInt(t, *PhotoConfInfo.Language, 1, "Language")
	assertFloatEqual(t, *PhotoConfInfo.Opacity, 0.75, "Opacity")
	assertEqual(t, *PhotoConfInfo.LogoTextContent, "Test Logo", "LogoTextContent")

	assertEqual(t, *RagConfInfo.EmbeddingType, "openai", "EmbeddingType")
	assertEqual(t, *RagConfInfo.KnowledgePath, "/data/knowledge", "KnowledgePath")
	assertEqual(t, *RagConfInfo.VectorDBType, "milvus", "VectorDBType")
	assertEqual(t, *RagConfInfo.ChromaURL, "http://localhost:8000", "ChromaURL")
	assertEqual(t, *RagConfInfo.Space, "test-space", "ChromaSpace")
	assertInt(t, *RagConfInfo.ChunkSize, 500, "ChunkSize")
	assertInt(t, *RagConfInfo.ChunkOverlap, 50, "ChunkOverlap")

	assertEqual(t, *McpConfPath, "./conf/mcp/mcp.json", "MCP_CONF_PATH")

	assertEqual(t, *VideoConfInfo.VolVideoModel, "model-v1", "VOL_VIDEO_MODEL")
	assertEqual(t, *VideoConfInfo.Radio, "radio-123", "RADIO")
	assertInt(t, *VideoConfInfo.Duration, 120, "DURATION")
	assertInt(t, *VideoConfInfo.FPS, 30, "FPS")
	assertEqual(t, *VideoConfInfo.Resolution, "1920x1080", "RESOLUTION")
	assertBool(t, *VideoConfInfo.Watermark, true, "WATERMARK")

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
