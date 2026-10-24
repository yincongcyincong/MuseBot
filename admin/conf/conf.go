package conf

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/yincongcyincong/MuseBot/logger"
	botUtils "github.com/yincongcyincong/MuseBot/utils"
)

// legacySessionKeyDefault was a hard-coded default; anyone can read it in the
// source or old docs and forge session cookies, so it must never be used.
const legacySessionKeyDefault = "muse_bot_session_key"

const sessionKeyFile = "data/admin_session_key"

type BaseConfig struct {
	DBType string `json:"db_type"`
	DBConf string `json:"db_conf"`

	SessionKey string `json:"session_key"`

	// SessionSecure marks the session cookie Secure so browsers only send it
	// over HTTPS; off by default because most deployments serve plain HTTP.
	SessionSecure bool `json:"session_secure"`

	AdminPort string `json:"admin_port"`

	CheckBotSec int `json:"check_bot_sec"`
}

var BaseConfInfo = new(BaseConfig)

func InitConfig() {
	flag.StringVar(&BaseConfInfo.DBType, "db_type", "sqlite3", "db type")
	flag.StringVar(&BaseConfInfo.DBConf, "db_conf", botUtils.GetAbsPath("data/muse_bot_admin.db"), "db conf")
	flag.StringVar(&BaseConfInfo.SessionKey, "session_key", "", "session key; auto-generated and persisted when empty")
	flag.BoolVar(&BaseConfInfo.SessionSecure, "session_secure", false, "mark the admin session cookie as Secure (HTTPS only)")
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

	if os.Getenv("SESSION_SECURE") == "true" {
		BaseConfInfo.SessionSecure = true
	}

	if os.Getenv("ADMIN_PORT") != "" {
		BaseConfInfo.AdminPort = os.Getenv("ADMIN_PORT")
	}

	if os.Getenv("CHECK_BOT_SEC") != "" {
		BaseConfInfo.CheckBotSec, _ = strconv.Atoi(os.Getenv("CHECK_BOT_SEC"))
	}

	sessionKeySource := ensureSessionKey()

	logger.Info("CONF", "DBType", BaseConfInfo.DBType)
	logger.Info("CONF", "DBConf", BaseConfInfo.DBConf)
	// never log the session key itself
	logger.Info("CONF", "SessionKey", sessionKeySource)
	logger.Info("CONF", "SessionSecure", BaseConfInfo.SessionSecure)
	logger.Info("CONF", "AdminPort", BaseConfInfo.AdminPort)
	logger.Info("CONF", "CheckBotSec", BaseConfInfo.CheckBotSec)

	EnvRegisterConf()
}

// ensureSessionKey falls back to a persisted random key when none is
// configured or when the legacy hard-coded default is detected. The key is
// stored under the data directory so sessions survive restarts.
func ensureSessionKey() string {
	if BaseConfInfo.SessionKey != "" && BaseConfInfo.SessionKey != legacySessionKeyDefault {
		return "configured"
	}

	return ensureSessionKeyAt(botUtils.GetAbsPath(sessionKeyFile))
}

// ensureSessionKeyAt loads or creates the persisted session key at keyPath.
func ensureSessionKeyAt(keyPath string) string {
	if data, err := os.ReadFile(keyPath); err == nil {
		if key := string(bytes.TrimSpace(data)); len(key) >= 32 {
			BaseConfInfo.SessionKey = key
			return "loaded from " + keyPath
		}
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		logger.Fatal("generate session key fail", "err", err)
		return ""
	}
	BaseConfInfo.SessionKey = hex.EncodeToString(buf)

	if err := os.MkdirAll(filepath.Dir(keyPath), 0755); err != nil {
		logger.Fatal("create data direction fail", "err", err)
		return ""
	}
	if err := os.WriteFile(keyPath, []byte(BaseConfInfo.SessionKey), 0600); err != nil {
		logger.Fatal("persist session key fail", "err", err)
		return ""
	}

	logger.Warn("SESSION_KEY not configured; a random key was generated and saved (set SESSION_KEY to override)",
		"file", keyPath)
	return "generated at " + keyPath
}
