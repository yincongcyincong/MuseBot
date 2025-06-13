package conf

import (
	"flag"
	"os"

	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

var (
	AudioAppID   *string
	AudioToken   *string
	AudioCluster *string
)

func InitAudioConf() {
	AudioAppID = flag.String("audio_app_id", "", "audio app id")
	AudioToken = flag.String("audio_token", "", "audio token")
	AudioCluster = flag.String("audio_cluster", "", "audio cluster")

}

func EnvAudioConf() {
	if os.Getenv("AUDIO_APP_ID") != "" {
		*AudioAppID = os.Getenv("AUDIO_APP_ID")
	}
	if os.Getenv("AUDIO_TOKEN") != "" {
		*AudioToken = os.Getenv("AUDIO_TOKEN")
	}
	if os.Getenv("AUDIO_CLUSTER") != "" {
		*AudioCluster = os.Getenv("AUDIO_CLUSTER")
	}

	logger.Info("AUDIO_CONF", "AUDIO_APP_ID", *AudioAppID)
	logger.Info("AUDIO_CONF", "AUDIO_TOKEN", *AudioToken)
	logger.Info("AUDIO_CONF", "AUDIO_CLUSTER", *AudioCluster)
}
