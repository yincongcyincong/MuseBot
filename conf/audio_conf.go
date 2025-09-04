package conf

import (
	"flag"
	"os"
	
	"github.com/yincongcyincong/MuseBot/logger"
)

type AudioConf struct {
	VolAudioAppID      *string `json:"vol_audio_app_id"`
	VolAudioToken      *string `json:"vol_audio_token"`
	VolAudioRecCluster *string `json:"vol_audio_cluster"`
	VolAudioVoiceType  *string `json:"vol_audio_voice_type"`
	VolAudioTTSCluster *string `json:"vol_audio_tts_cluster"`
	
	GeminiAudioModel *string `json:"gemini_audio_model"`
	GeminiVoiceName  *string `json:"gemini_voice_name"`
	
	TTSType *string `json:"vol_tts_type"`
}

var (
	AudioConfInfo = new(AudioConf)
)

func InitAudioConf() {
	AudioConfInfo.VolAudioAppID = flag.String("vol_audio_app_id", "", "vol audio app id")
	AudioConfInfo.VolAudioToken = flag.String("vol_audio_token", "", "vol audio token")
	AudioConfInfo.VolAudioRecCluster = flag.String("vol_audio_rec_cluster", "volcengine_input_common", "vol audio cluster")
	AudioConfInfo.VolAudioVoiceType = flag.String("vol_audio_voice_type", "", "vol audio voice type")
	AudioConfInfo.VolAudioTTSCluster = flag.String("vol_audio_tts_cluster", "volcano_tts", "vol audio tts cluster")
	
	AudioConfInfo.GeminiAudioModel = flag.String("gemini_audio_model", "gemini-2.5-flash-preview-tts", "gemini audio model")
	AudioConfInfo.GeminiVoiceName = flag.String("gemini_voice_name", "Kore", "gemini voice name")
	
	AudioConfInfo.TTSType = flag.String("tts_type", "", "vol tts type: 1. vol 2. gemini")
}

func EnvAudioConf() {
	if os.Getenv("VOL_AUDIO_APP_ID") != "" {
		*AudioConfInfo.VolAudioAppID = os.Getenv("VOL_AUDIO_APP_ID")
	}
	if os.Getenv("VOL_AUDIO_TOKEN") != "" {
		*AudioConfInfo.VolAudioToken = os.Getenv("VOL_AUDIO_TOKEN")
	}
	if os.Getenv("VOL_AUDIO_REC_CLUSTER") != "" {
		*AudioConfInfo.VolAudioRecCluster = os.Getenv("VOL_AUDIO_REC_CLUSTER")
	}
	if os.Getenv("VOL_AUDIO_VOICE_TYPE") != "" {
		*AudioConfInfo.VolAudioVoiceType = os.Getenv("VOL_AUDIO_VOICE_TYPE")
	}
	
	if os.Getenv("VOL_AUDIO_TTS_CLUSTER") != "" {
		*AudioConfInfo.VolAudioTTSCluster = os.Getenv("VOL_AUDIO_TTS_CLUSTER")
	}
	
	if os.Getenv("GEMINI_AUDIO_MODEL") != "" {
		*AudioConfInfo.GeminiAudioModel = os.Getenv("GEMINI_AUDIO_MODEL")
	}
	
	if os.Getenv("GEMINI_VOICE_NAME") != "" {
		*AudioConfInfo.GeminiVoiceName = os.Getenv("GEMINI_VOICE_NAME")
	}
	
	if os.Getenv("TTS_TYPE") != "" {
		*AudioConfInfo.TTSType = os.Getenv("TTS_TYPE")
	}
	
	logger.Info("AUDIO_CONF", "AudioAppID", *AudioConfInfo.VolAudioAppID)
	logger.Info("AUDIO_CONF", "AudioToken", *AudioConfInfo.VolAudioToken)
	logger.Info("AUDIO_CONF", "AudioCluster", *AudioConfInfo.VolAudioRecCluster)
	logger.Info("AUDIO_CONF", "AudioVoiceType", *AudioConfInfo.VolAudioVoiceType)
	logger.Info("AUDIO_CONF", "AudioTTSCluster", *AudioConfInfo.VolAudioTTSCluster)
	logger.Info("AUDIO_CONF", "GeminiAudioModel", *AudioConfInfo.GeminiAudioModel)
	logger.Info("AUDIO_CONF", "GeminiVoiceName", *AudioConfInfo.GeminiVoiceName)
	logger.Info("AUDIO_CONF", "TTSType", *AudioConfInfo.TTSType)
}
