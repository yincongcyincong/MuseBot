package conf

import (
	"flag"
	"os"
	"strconv"
)

type AudioConf struct {
	VolAudioAppID      string `json:"vol_audio_app_id"`
	VolAudioToken      string `json:"vol_audio_token"`
	VolAudioRecCluster string `json:"vol_audio_rec_cluster"`
	VolAudioVoiceType  string `json:"vol_audio_voice_type"`
	VolAudioTTSCluster string `json:"vol_audio_tts_cluster"`
	VolEndSmoothWindow int    `json:"vol_end_smooth_window"`
	VolTTSSpeaker      string `json:"vol_tts_speaker"`
	VolBotName         string `json:"vol_bot_name"`
	VolSystemRole      string `json:"vol_system_role"`
	VolSpeakingStyle   string `json:"vol_speaking_style"`
	
	GeminiAudioModel string `json:"gemini_audio_model"`
	GeminiVoiceName  string `json:"gemini_voice_name"`
	
	OpenAIAudioModel string `json:"openai_audio_model"`
	OpenAIVoiceName  string `json:"openai_voice_name"`
	
	AliyunAudioModel    string `json:"aliyun_audio_model"`
	AliyunAudioVoice    string `json:"aliyun_audio_voice"`
	AliyunAudioRecModel string `json:"aliyun_audio_rec_model"`
	
	TTSType string `json:"tts_type"`
}

var (
	AudioConfInfo = new(AudioConf)
)

func InitAudioConf() {
	flag.StringVar(&AudioConfInfo.VolAudioAppID, "vol_audio_app_id", "", "vol audio app id")
	flag.StringVar(&AudioConfInfo.VolAudioToken, "vol_audio_token", "", "vol audio token")
	flag.StringVar(&AudioConfInfo.VolAudioRecCluster, "vol_audio_rec_cluster", "volcengine_input_common", "vol audio cluster")
	flag.StringVar(&AudioConfInfo.VolAudioVoiceType, "vol_audio_voice_type", "", "vol audio voice type")
	flag.StringVar(&AudioConfInfo.VolAudioTTSCluster, "vol_audio_tts_cluster", "volcano_tts", "vol audio tts cluster")
	flag.IntVar(&AudioConfInfo.VolEndSmoothWindow, "vol_end_smooth_window", 1500, "vol end smooth window")
	flag.StringVar(&AudioConfInfo.VolTTSSpeaker, "vol_tts_speaker", "zh_female_vv_jupiter_bigtts", "vol tts speaker")
	flag.StringVar(&AudioConfInfo.VolBotName, "vol_bot_name", "豆包", "vol bot name")
	flag.StringVar(&AudioConfInfo.VolSystemRole, "vol_system_role", "你使用活泼灵动的女声，性格开朗，热爱生活。", "vol system role")
	flag.StringVar(&AudioConfInfo.VolSpeakingStyle, "vol_speaking_style", "你的说话风格简洁明了，语速适中，语调自然。", "vol speaking style")
	
	flag.StringVar(&AudioConfInfo.GeminiAudioModel, "gemini_audio_model", "gemini-2.5-flash-preview-tts", "gemini audio model")
	flag.StringVar(&AudioConfInfo.GeminiVoiceName, "gemini_voice_name", "Kore", "gemini voice name")
	
	flag.StringVar(&AudioConfInfo.OpenAIAudioModel, "openai_audio_model", "tts-1", "openai audio model")
	flag.StringVar(&AudioConfInfo.OpenAIVoiceName, "openai_voice_name", "alloy", "openai voice name")
	
	flag.StringVar(&AudioConfInfo.AliyunAudioModel, "aliyun_audio_model", "qwen3-tts-flash", "aliyun audio model")
	flag.StringVar(&AudioConfInfo.AliyunAudioVoice, "aliyun_audio_voice", "Cherry", "aliyun audio voice")
	flag.StringVar(&AudioConfInfo.AliyunAudioRecModel, "aliyun_audio_rec_model", "qwen-audio-turbo-latest", "aliyun audio rec model")
	
	flag.StringVar(&AudioConfInfo.TTSType, "tts_type", "", "vol tts type: 1. vol 2. gemini")
}

func EnvAudioConf() {
	if os.Getenv("VOL_AUDIO_APP_ID") != "" {
		AudioConfInfo.VolAudioAppID = os.Getenv("VOL_AUDIO_APP_ID")
	}
	if os.Getenv("VOL_AUDIO_TOKEN") != "" {
		AudioConfInfo.VolAudioToken = os.Getenv("VOL_AUDIO_TOKEN")
	}
	if os.Getenv("VOL_AUDIO_REC_CLUSTER") != "" {
		AudioConfInfo.VolAudioRecCluster = os.Getenv("VOL_AUDIO_REC_CLUSTER")
	}
	if os.Getenv("VOL_AUDIO_VOICE_TYPE") != "" {
		AudioConfInfo.VolAudioVoiceType = os.Getenv("VOL_AUDIO_VOICE_TYPE")
	}
	
	if os.Getenv("VOL_AUDIO_TTS_CLUSTER") != "" {
		AudioConfInfo.VolAudioTTSCluster = os.Getenv("VOL_AUDIO_TTS_CLUSTER")
	}
	
	if os.Getenv("GEMINI_AUDIO_MODEL") != "" {
		AudioConfInfo.GeminiAudioModel = os.Getenv("GEMINI_AUDIO_MODEL")
	}
	
	if os.Getenv("GEMINI_VOICE_NAME") != "" {
		AudioConfInfo.GeminiVoiceName = os.Getenv("GEMINI_VOICE_NAME")
	}
	
	if os.Getenv("OPENAI_AUDIO_MODEL") != "" {
		AudioConfInfo.OpenAIAudioModel = os.Getenv("OPENAI_AUDIO_MODEL")
	}
	
	if os.Getenv("OPENAI_VOICE_NAME") != "" {
		AudioConfInfo.OpenAIVoiceName = os.Getenv("OPENAI_VOICE_NAME")
	}
	
	if os.Getenv("TTS_TYPE") != "" {
		AudioConfInfo.TTSType = os.Getenv("TTS_TYPE")
	}
	
	if os.Getenv("VOL_END_SMOOTH_WINDOW") != "" {
		AudioConfInfo.VolEndSmoothWindow, _ = strconv.Atoi(os.Getenv("VOL_END_SMOOTH_WINDOW"))
	}
	
	if os.Getenv("VOL_TTS_SPEAKER") != "" {
		AudioConfInfo.VolTTSSpeaker = os.Getenv("VOL_TTS_SPEAKER")
	}
	
	if os.Getenv("VOL_BOT_NAME") != "" {
		AudioConfInfo.VolBotName = os.Getenv("VOL_BOT_NAME")
	}
	
	if os.Getenv("VOL_SYSTEM_ROLE") != "" {
		AudioConfInfo.VolSystemRole = os.Getenv("VOL_SYSTEM_ROLE")
	}
	
	if os.Getenv("VOL_SPEAKING_STYLE") != "" {
		AudioConfInfo.VolSpeakingStyle = os.Getenv("VOL_SPEAKING_STYLE")
	}
	
	if os.Getenv("ALIYUN_AUDIO_MODEL") != "" {
		AudioConfInfo.AliyunAudioModel = os.Getenv("ALIYUN_AUDIO_MODEL")
	}
	
	if os.Getenv("ALIYUN_AUDIO_VOICE") != "" {
		AudioConfInfo.AliyunAudioVoice = os.Getenv("ALIYUN_AUDIO_VOICE")
	}
	
	if os.Getenv("ALIYUN_AUDIO_REC_MODEL") != "" {
		AudioConfInfo.AliyunAudioRecModel = os.Getenv("ALIYUN_AUDIO_REC_MODEL")
	}
	
}
