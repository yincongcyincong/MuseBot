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
	os.Setenv("DB_TYPE", "sqlite3")
	os.Setenv("DB_CONF", "./test.db")
	os.Setenv("ALLOWED_TELEGRAM_USER_IDS", "0")
	os.Setenv("ALLOWED_TELEGRAM_GROUP_IDS", "0")
	os.Setenv("DEEPSEEK_PROXY", "http://127.0.0.1:7890")
	os.Setenv("TELEGRAM_PROXY", "http://127.0.0.1:7891")

	// call InitConf
	InitConf()

	// assertion testing
	assert.Equal(t, "test_bot_token", *BotToken, "BotToken should be from environment variable")
	assert.Equal(t, "test_deepseek_token", *DeepseekToken, "DeepseekToken should be from environment variable")
	assert.Equal(t, "https://test.deepseek.com/", *CustomUrl, "CustomUrl should be from environment variable")
	assert.Equal(t, "test_type", *DeepseekType, "DeepseekType should be from environment variable")
	assert.Equal(t, "ak", *VolcAK, "VolcAK should be from environment variable")
	assert.Equal(t, "sk", *VolcSK, "VolcSK should be from environment variable")
	assert.Equal(t, "sqlite3", *DBType, "DBType should be from environment variable")
	assert.Equal(t, "./test.db", *DBConf, "DBPath should be from environment variable")
	assert.Equal(t, "http://127.0.0.1:7891", *TelegramProxy, "TelegramProxy should be from environment variable")
	assert.Equal(t, map[int64]bool{0: true}, AllowedTelegramUserIds, "AllowedTelegramUserIds should be from environment variable")
	assert.Equal(t, map[int64]bool{0: true}, AllowedTelegramGroupIds, "AllowedTelegramGroupIds should be from environment variable")
	assert.Equal(t, "http://127.0.0.1:7890", *DeepseekProxy, "DeepseekProxy should be from environment variable")

}
