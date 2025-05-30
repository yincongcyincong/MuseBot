package conf

import (
	"flag"
	"os"
	"testing"
)

func TestInitConf_InitTools(t *testing.T) {
	if UseTools == nil {
		UseTools = getPointBool(true)
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

		InitToolsConf()
		flag.Parse()
	}

	InitTools()
	if len(DeepseekTools) != len(OpenAITools) {
		t.Errorf("%s expected %d, got %d", "tools number", len(DeepseekTools), len(OpenAITools))
	}
}

func getPointBool(b bool) *bool {
	return &b
}
