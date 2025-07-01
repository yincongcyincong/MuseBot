package conf

import (
	"flag"
	"os"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type AudioConf struct {
	AudioAppID   *string
	AudioToken   *string
	AudioCluster *string
}

var (
	AudioConfInfo = new(AudioConf)
)

func InitAudioConf() {
	AudioConfInfo.AudioAppID = flag.String("audio_app_id", "", "audio app id")
	AudioConfInfo.AudioToken = flag.String("audio_token", "", "audio token")
	AudioConfInfo.AudioCluster = flag.String("audio_cluster", "", "audio cluster")
	
}

func EnvAudioConf() {
	if os.Getenv("AUDIO_APP_ID") != "" {
		*AudioConfInfo.AudioAppID = os.Getenv("AUDIO_APP_ID")
	}
	if os.Getenv("AUDIO_TOKEN") != "" {
		*AudioConfInfo.AudioToken = os.Getenv("AUDIO_TOKEN")
	}
	if os.Getenv("AUDIO_CLUSTER") != "" {
		*AudioConfInfo.AudioCluster = os.Getenv("AUDIO_CLUSTER")
	}
	
	logger.Info("AUDIO_CONF", "AUDIO_APP_ID", *AudioConfInfo.AudioAppID)
	logger.Info("AUDIO_CONF", "AUDIO_TOKEN", *AudioConfInfo.AudioToken)
	logger.Info("AUDIO_CONF", "AUDIO_CLUSTER", *AudioConfInfo.AudioCluster)
}
