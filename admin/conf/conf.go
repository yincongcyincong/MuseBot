package conf

import (
	"flag"
	"os"
	
	"github.com/yincongcyincong/MuseBot/logger"
)

type BaseConfig struct {
	DBType *string `json:"db_type"`
	DBConf *string `json:"db_conf"`
	
	SessionKey *string `json:"session_key"`
	
	AdminPort *string `json:"admin_port"`
}

var BaseConfInfo = new(BaseConfig)

func InitConfig() {
	BaseConfInfo.DBType = flag.String("db_type", "sqlite3", "db type")
	BaseConfInfo.DBConf = flag.String("db_conf", "./data/telegram_admin_bot.db", "db conf")
	BaseConfInfo.SessionKey = flag.String("session_key", "telegram_bot_session_key", "session key")
	BaseConfInfo.AdminPort = flag.String("admin_port", "18080", "admin port")
	
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
	
	if os.Getenv("ADMIN_PORT") != "" {
		*BaseConfInfo.AdminPort = os.Getenv("ADMIN_PORT")
	}
	
	logger.Info("CONF", "DBType", *BaseConfInfo.DBType)
	logger.Info("CONF", "DBConf", *BaseConfInfo.DBConf)
	logger.Info("CONF", "SessionKey", *BaseConfInfo.SessionKey)
	logger.Info("CONF", "AdminPort", *BaseConfInfo.AdminPort)
}
