package conf

import (
	"flag"
	"os"
	
	"github.com/yincongcyincong/MuseBot/logger"
)

type AudioConf struct {
	VolAudioAppID      *string `json:"vol_audio_app_id"`
	VolAudioToken      *string `json:"vol_audio_token"`
	VolAudioRecCluster *string `json:"vol_audio_rec_cluster"`
	VolAudioVoiceType  *string `json:"vol_audio_voice_type"`
	VolAudioTTSCluster *string `json:"vol_audio_tts_cluster"`
	VolEndSmoothWindow *int    `json:"vol_end_smooth_window"`
	VolTTSSpeaker      *string `json:"vol_tts_speaker"`
	VolBotName         *string `json:"vol_bot_name"`
	VolSystemRole      *string `json:"vol_system_role"`
	VolSpeakingStyle   *string `json:"vol_speaking_style"`
	
	GeminiAudioModel *string `json:"gemini_audio_model"`
	GeminiVoiceName  *string `json:"gemini_voice_name"`
	
	OpenAIAudioModel *string `json:"openai_audio_model"`
	OpenAIVoiceName  *string `json:"openai_voice_name"`
	
	AliyunAudioModel    *string `json:"aliyun_audio_model"`
	AliyunAudioVoice    *string `json:"aliyun_audio_voice"`
	AliyunAudioRecModel *string `json:"aliyun_audio_rec_model"`
	
	TTSType *string `json:"tts_type"`
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
	AudioConfInfo.VolEndSmoothWindow = flag.Int("vol_end_smooth_window", 1500, "vol end smooth window")
	AudioConfInfo.VolTTSSpeaker = flag.String("vol_tts_speaker", "zh_female_vv_jupiter_bigtts", "vol tts speaker")
	AudioConfInfo.VolBotName = flag.String("vol_bot_name", "豆包", "vol bot name")
	AudioConfInfo.VolSystemRole = flag.String("vol_system_role", "你使用活泼灵动的女声，性格开朗，热爱生活。", "vol system role")
	AudioConfInfo.VolSpeakingStyle = flag.String("vol_speaking_style", "你的说话风格简洁明了，语速适中，语调自然。", "vol speaking style")
	
	AudioConfInfo.GeminiAudioModel = flag.String("gemini_audio_model", "gemini-2.5-flash-preview-tts", "gemini audio model")
	AudioConfInfo.GeminiVoiceName = flag.String("gemini_voice_name", "Kore", "gemini voice name")
	
	AudioConfInfo.OpenAIAudioModel = flag.String("openai_audio_model", "tts-1", "openai audio model")
	AudioConfInfo.OpenAIVoiceName = flag.String("openai_voice_name", "alloy", "openai voice name")
	
	AudioConfInfo.AliyunAudioModel = flag.String("aliyun_audio_model", "qwen3-tts-flash", "aliyun audio model")
	AudioConfInfo.AliyunAudioVoice = flag.String("aliyun_audio_voice", "Cherry", "aliyun audio voice")
	AudioConfInfo.AliyunAudioRecModel = flag.String("aliyun_audio_rec_model", "qwen-audio-turbo-latest", "aliyun audio rec model")
	
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
	
	if os.Getenv("OPENAI_AUDIO_MODEL") != "" {
		*AudioConfInfo.OpenAIAudioModel = os.Getenv("OPENAI_AUDIO_MODEL")
	}
	
	if os.Getenv("OPENAI_VOICE_NAME") != "" {
		*AudioConfInfo.OpenAIVoiceName = os.Getenv("OPENAI_VOICE_NAME")
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
	logger.Info("AUDIO_CONF", "OpenAIAudioModel", *AudioConfInfo.OpenAIAudioModel)
	logger.Info("AUDIO_CONF", "OpenAIVoiceName", *AudioConfInfo.OpenAIVoiceName)
	logger.Info("AUDIO_CONF", "TTSType", *AudioConfInfo.TTSType)
	
	logger.Info("AUDIO_CONF", "VolEndSmoothWindow", *AudioConfInfo.VolEndSmoothWindow)
	logger.Info("AUDIO_CONF", "VolTTSSpeaker", *AudioConfInfo.VolTTSSpeaker)
	logger.Info("AUDIO_CONF", "VolBotName", *AudioConfInfo.VolBotName)
	logger.Info("AUDIO_CONF", "VolSystemRole", *AudioConfInfo.VolSystemRole)
	logger.Info("AUDIO_CONF", "VolSpeakingStyle", *AudioConfInfo.VolSpeakingStyle)
	logger.Info("AUDIO_CONF", "AliyunAudioModel", *AudioConfInfo.AliyunAudioModel)
	logger.Info("AUDIO_CONF", "AliyunAudioVoice", *AudioConfInfo.AliyunAudioVoice)
	logger.Info("AUDIO_CONF", "AliyunAudioRecModel", *AudioConfInfo.AliyunAudioRecModel)
	
}
