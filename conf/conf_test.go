package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConf(t *testing.T) {
	// set env
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	os.Setenv("CUSTOM_URL", "https://test.deepseek.com/")
	os.Setenv("DEEPSEEK_TYPE", "test_type")
	os.Setenv("VOLC_AK", "ak")
	os.Setenv("VOLC_SK", "sk")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DBType", "sqlite3")
	os.Setenv("DBPath", "./test.db")

	// call InitConf
	InitConf()

	// assertion testing
	assert.Equal(t, "test_bot_token", *BotToken, "BotToken should be from environment variable")
	assert.Equal(t, "test_deepseek_token", *DeepseekToken, "DeepseekToken should be from environment variable")
	assert.Equal(t, "https://test.deepseek.com/", *CustomUrl, "CustomUrl should be from environment variable")
	assert.Equal(t, "test_type", *DeepseekType, "DeepseekType should be from environment variable")
	assert.Equal(t, "ak", *VolcAK, "VolcAK should be from environment variable")
	assert.Equal(t, "sk", *VolcSK, "VolcSK should be from environment variable")

}
