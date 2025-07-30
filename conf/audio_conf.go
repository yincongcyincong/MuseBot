package conf

import (
	"flag"
	"os"
	"strconv"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type AudioConf struct {
	AudioAppID   *string `json:"audio_app_id"`
	AudioToken   *string `json:"audio_token"`
	AudioCluster *string `json:"audio_cluster"`
	
	GeminiVideoModel            *string `json:"gemini_video_model"`
	GeminiVideoAspectRatio      *string `json:"gemini_video_aspect_ratio"`
	GeminiVideoDurationSeconds  int32   `json:"gemini_video_duration_seconds"`
	GeminiVideoPersonGeneration *string `json:"gemini_video_person_generation"`
}

var (
	AudioConfInfo = new(AudioConf)
)

func InitAudioConf() {
	AudioConfInfo.AudioAppID = flag.String("audio_app_id", "", "audio app id")
	AudioConfInfo.AudioToken = flag.String("audio_token", "", "audio token")
	AudioConfInfo.AudioCluster = flag.String("audio_cluster", "", "audio cluster")
	
	AudioConfInfo.GeminiVideoModel = flag.String("gemini_video_model", "veo-2.0-generate-001", "create video model")
	AudioConfInfo.GeminiVideoAspectRatio = flag.String("gemini_video_aspect_ratio", "16:9", "gemini video ratio")
	AudioConfInfo.GeminiVideoDurationSeconds = int32(*flag.Int("gemini_video_duration_seconds", 0, "gemini video duration"))
	AudioConfInfo.GeminiVideoPersonGeneration = flag.String("gemini_video_person_generation", "allow_all", "gemini video can generate person or not: allow_all or ")
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
	
	if os.Getenv("GEMINI_VIDEO_MODEL") != "" {
		*AudioConfInfo.GeminiVideoModel = os.Getenv("GEMINI_VIDEO_MODEL")
	}
	if os.Getenv("GEMINI_VIDEO_PERSON_GENERATION") != "" {
		*AudioConfInfo.GeminiVideoPersonGeneration = os.Getenv("GEMINI_VIDEO_PERSON_GENERATION")
	}
	if os.Getenv("GEMINI_VIDEO_ASPECT_RATIO") != "" {
		*AudioConfInfo.GeminiVideoAspectRatio = os.Getenv("GEMINI_VIDEO_ASPECT_RATIO")
	}
	if os.Getenv("GEMINI_VIDEO_DURATION_SECONDS") != "" {
		tmp, _ := strconv.Atoi(os.Getenv("GEMINI_VIDEO_DURATION_SECONDS"))
		AudioConfInfo.GeminiVideoDurationSeconds = int32(tmp)
	}
	
	logger.Info("AUDIO_CONF", "AudioAppID", *AudioConfInfo.AudioAppID)
	logger.Info("AUDIO_CONF", "AudioToken", *AudioConfInfo.AudioToken)
	logger.Info("AUDIO_CONF", "AudioCluster", *AudioConfInfo.AudioCluster)
	logger.Info("AUDIO_CONF", "GeminiVideoModel", *AudioConfInfo.GeminiVideoModel)
	logger.Info("AUDIO_CONF", "GeminiVideoPersonGeneration", *AudioConfInfo.GeminiVideoPersonGeneration)
	logger.Info("AUDIO_CONF", "GeminiVideoAspectRatio", *AudioConfInfo.GeminiVideoAspectRatio)
	logger.Info("AUDIO_CONF", "GeminiVideoDurationSeconds", AudioConfInfo.GeminiVideoDurationSeconds)
}
