package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConf(t *testing.T) {
	// 设置环境变量
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("CUSTOM_URL", "https://test.deepseek.com/")
	os.Setenv("DEEPSEEK_TYPE", "test_type")

	// 调用 InitConf
	InitConf()

	// 断言测试
	assert.Equal(t, "test_bot_token", *BotToken, "BotToken should be from environment variable")
	assert.Equal(t, "test_deepseek_token", *DeepseekToken, "DeepseekToken should be from environment variable")
	assert.Equal(t, "https://test.deepseek.com/", *CustomUrl, "CustomUrl should be from environment variable")
	assert.Equal(t, "test_type", *DeepseekType, "DeepseekType should be from environment variable")
}
