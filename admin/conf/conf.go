package conf

import (
	"flag"
	"os"
	"strconv"

	"github.com/yincongcyincong/MuseBot/logger"
	botUtils "github.com/yincongcyincong/MuseBot/utils"
)

type BaseConfig struct {
	DBType *string `json:"db_type"`
	DBConf *string `json:"db_conf"`

	SessionKey *string `json:"session_key"`

	AdminPort *string `json:"admin_port"`

	CheckBotSec *int `json:"check_bot_sec"`
}

var BaseConfInfo = new(BaseConfig)

func InitConfig() {
	BaseConfInfo.DBType = flag.String("db_type", "sqlite3", "db type")
	BaseConfInfo.DBConf = flag.String("db_conf", botUtils.GetAbsPath("data/muse_bot_admin_bot.db"), "db conf")
	BaseConfInfo.SessionKey = flag.String("session_key", "muse_bot_session_key", "session key")
	BaseConfInfo.AdminPort = flag.String("admin_port", "18080", "admin port")
	BaseConfInfo.CheckBotSec = flag.Int("check_bot_sec", 10, "check bot interval")

	InitRegisterConf()
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

	if os.Getenv("CHECK_BOT_SEC") != "" {
		*BaseConfInfo.CheckBotSec, _ = strconv.Atoi(os.Getenv("CHECK_BOT_SEC"))
	}

	logger.Info("CONF", "DBType", *BaseConfInfo.DBType)
	logger.Info("CONF", "DBConf", *BaseConfInfo.DBConf)
	logger.Info("CONF", "SessionKey", *BaseConfInfo.SessionKey)
	logger.Info("CONF", "AdminPort", *BaseConfInfo.AdminPort)
	logger.Info("CONF", "CheckBotSec", *BaseConfInfo.CheckBotSec)

	EnvRegisterConf()
}
