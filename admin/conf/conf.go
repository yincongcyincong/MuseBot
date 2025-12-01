package conf

import (
	"flag"
	"os"
	"strconv"
	
	"github.com/yincongcyincong/MuseBot/logger"
	botUtils "github.com/yincongcyincong/MuseBot/utils"
)

type BaseConfig struct {
	DBType string `json:"db_type"`
	DBConf string `json:"db_conf"`
	
	SessionKey string `json:"session_key"`
	
	AdminPort string `json:"admin_port"`
	
	CheckBotSec int `json:"check_bot_sec"`
}

var BaseConfInfo = new(BaseConfig)

func InitConfig() {
	flag.StringVar(&BaseConfInfo.DBType, "db_type", "sqlite3", "db type")
	flag.StringVar(&BaseConfInfo.DBConf, "db_conf", botUtils.GetAbsPath("data/muse_bot_admin.db"), "db conf")
	flag.StringVar(&BaseConfInfo.SessionKey, "session_key", "muse_bot_session_key", "session key")
	flag.StringVar(&BaseConfInfo.AdminPort, "admin_port", "18080", "admin port")
	flag.IntVar(&BaseConfInfo.CheckBotSec, "check_bot_sec", 10, "check bot interval")
	
	InitRegisterConf()
	
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.Parse()
	
	if os.Getenv("DB_TYPE") != "" {
		BaseConfInfo.DBType = os.Getenv("DB_TYPE")
	}
	
	if os.Getenv("DB_CONF") != "" {
		BaseConfInfo.DBConf = os.Getenv("DB_CONF")
	}
	
	if os.Getenv("SESSION_KEY") != "" {
		BaseConfInfo.SessionKey = os.Getenv("SESSION_KEY")
	}
	
	if os.Getenv("ADMIN_PORT") != "" {
		BaseConfInfo.AdminPort = os.Getenv("ADMIN_PORT")
	}
	
	if os.Getenv("CHECK_BOT_SEC") != "" {
		BaseConfInfo.CheckBotSec, _ = strconv.Atoi(os.Getenv("CHECK_BOT_SEC"))
	}
	
	logger.Info("CONF", "DBType", BaseConfInfo.DBType)
	logger.Info("CONF", "DBConf", BaseConfInfo.DBConf)
	logger.Info("CONF", "SessionKey", BaseConfInfo.SessionKey)
	logger.Info("CONF", "AdminPort", BaseConfInfo.AdminPort)
	logger.Info("CONF", "CheckBotSec", BaseConfInfo.CheckBotSec)
	
	EnvRegisterConf()
}
