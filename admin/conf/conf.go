package conf

import (
	"flag"
	"os"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type BaseConfig struct {
	DBType *string `json:"db_type"`
	DBConf *string `json:"db_conf"`
	
	SessionKey *string `json:"session_key"`
}

var BaseConfInfo = new(BaseConfig)

func InitConfig() {
	BaseConfInfo.DBType = flag.String("db_type", "sqlite3", "db type")
	BaseConfInfo.DBConf = flag.String("db_conf", "./data/telegram_admin_bot.db", "db conf")
	BaseConfInfo.SessionKey = flag.String("session_key", "telegram_bot_session_key", "session key")
	
	flag.Parse()
	
	if os.Getenv("DB_TYPE") != "" {
		*BaseConfInfo.DBType = os.Getenv("DB_TYPE")
	}
	
	if os.Getenv("DB_CONF") != "" {
		*BaseConfInfo.DBConf = os.Getenv("DB_CONF")
	}
	
	if os.Getenv("SESSION_KEY") != "" {
		*BaseConfInfo.SessionKey = os.Getenv("SESSION_KEY")
	}
	
	logger.Info("CONF", "DBType", *BaseConfInfo.DBType)
	logger.Info("CONF", "DBConf", *BaseConfInfo.DBConf)
	logger.Info("CONF", "SessionKey", *BaseConfInfo.SessionKey)
}
