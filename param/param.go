package param

import (
	"github.com/cohesion-org/deepseek-go"
	"github.com/devinyf/dashscopego/qwen"
)

const (
	DeepSeek     = "deepseek"
	Ollama       = "ollama"
	ChatAnyWhere = "chatanywhere"
	
	Vol    = "vol"
	AI302  = "302-ai"
	Aliyun = "aliyun"
	
	Gemini                        = "gemini"
	ModelGemini25Pro       string = "gemini-2.5-pro"
	ModelGemini25Flash     string = "gemini-2.5-flash"
	ModelGemini20Flash     string = "gemini-2.0-flash"
	ModelGemini20FlashLite string = "gemini-2.0-flash-lite"
	ModelGemini15Pro       string = "gemini-1.5-pro"
	ModelGemini15Flash     string = "gemini-1.5-flash"
	
	// 特定功能模型
	ModelGeminiFlashPreviewTTS string = "gemini-flash-preview-tts"
	ModelGeminiEmbedding       string = "gemini-embedding"
	ModelImagen3               string = "imagen-3"
	ModelVeo2                  string = "veo-2"
	
	OpenAi = "openai"
	
	OpenRouter = "openrouter"
	
	LLAVA = "llava:latest"
	
	DiscordEditMode = "edit"
	
	ImageTokenUsage = 3000
	AudioTokenUsage = 500
	VideoTokenUsage = 5000
)

const (
	// doubao Seed 1.6
	ModelDoubaoSeed16         = "doubao-seed-1.6-250615"
	ModelDoubaoSeed16Flash    = "doubao-seed-1.6-flash-250615"
	ModelDoubaoSeed16Thinking = "doubao-seed-1.6-thinking-250615"
	
	// doubao 1.5 系列
	ModelDoubao15ThinkingPro     = "doubao-1.5-thinking-pro-250415"
	ModelDoubao15ThinkingProM428 = "doubao-1.5-thinking-pro-m-250428"
	ModelDoubao15ThinkingProM415 = "doubao-1.5-thinking-pro-m-250415"
	ModelDoubao15VisionPro428    = "doubao-1.5-thinking-vision-pro-250428"
	ModelDoubao15VisionPro328    = "doubao-1.5-vision-pro-250328"
	
	// DeepSeek R1
	ModelDeepSeekR1        = "deepseek-r1"
	ModelDeepSeekR1_528    = "deepseek-r1-250528"
	ModelDeepSeekR1_120    = "deepseek-r1-250120"
	ModelDeepSeekR1Qwen32b = "deepseek-r1-distill-qwen-32b-250120"
	ModelDeepSeekR1Qwen7b  = "deepseek-r1-distill-qwen-7b-250120"
	
	ModelDeepseekV3         = "deepseek-v3.1"
	ModelDeepSeekR1Qwen32bQ = "deepseek-r1-distill-qwen-32b"
	ModelDeepSeekR1Qwen7bQ  = "deepseek-r1-distill-qwen-7b"
	
	// doubao-1.5
	ModelDoubao15VisionPro32k = "doubao-1.5-vision-pro-32k-250115"
	ModelDoubao15VisionLite   = "doubao-1.5-vision-lite-250315"
	
	AllRecordType   = -1
	TextRecordType  = 0
	ImageRecordType = 1
	VideoRecordType = 2
	WEBRecordType   = 3
	TalkRecordType  = 4
	
	DefaultContextToken = 128000
	
	Gemini2_5FlashPreview    = "gemini-2.5-flash-image-preview"
	GeminiImageGenV2_5       = "gemini-2.5-flash-image"
	Imagen3_0Generate002     = "imagen-3.0-generate-002"
	Imagen4_0Generate001     = "imagen-4.0-generate-001"
	Imagen4_0FastGenerate002 = "imagen-4.0-fast-generate-001"
	
	GeminiVideoVeo3_1Preview     = "veo-3.1-generate-preview"
	GeminiVideoVeo3_1FastPreview = "veo-3.1-fast-generate-preview"
	GeminiVideoVeo2              = "veo-2.0-generate-001"
	GeminiVideoVeo3              = "veo-3.0-generate-001"
	GeminiVideoFastVeo3          = "veo-3.0-fast-generate-001"
	
	ModelImageGPT = "gpt-image-1"
	
	DoubaoSeed16VisionPro = "doubao-seed-1-6-250615"
	
	QwenImagePlus   = "qwen-image-plus"
	QwenVlMaxLatest = "qwen-vl-max-latest"
	QwenVlMax       = "qwen-vl-max"
	QVQMax          = "qvq-max"
	QVQMaxLatest    = "qvq-max-latest"
	
	QwenAudioTurbo = "qwen-audio-turbo-latest"
	
	Wan2_5T2VPreview = "wan2.5-t2v-preview"
	
	DoubaoSeedance1_0Pro = "doubao-seedance-1-0-pro-250528"
	
	ChatGPT4_0 = "chatgpt-4o-latest"
	
	VolTTS = "volcano_tts"
	VolIcl = "volcano_icl"
	
	Qwen3TTSFlash = "qwen3-tts-flash"
	
	Gemini2_5FlashPreviewTTS = "gemini-2.5-flash-preview-tts"
	
	ComWechat  = "com_wechat"
	Ding       = "ding"
	Discord    = "discord"
	Lark       = "lark"
	PersonalQQ = "personal_qq"
	QQ         = "qq"
	Slack      = "slack"
	Telegram   = "telegram"
	Wechat     = "wechat"
	
	State      = "state"
	Clear      = "clear"
	Retry      = "retry"
	Chat       = "chat"
	Photo      = "photo"
	EditPhoto  = "edit_photo"
	Video      = "video"
	Help       = "help"
	Task       = "task"
	Mcp        = "mcp"
	Mode       = "mode"
	TxtType    = "txt_type"
	TxtModel   = "txt_model"
	PhotoType  = "photo_type"
	PhotoModel = "photo_model"
	VideoType  = "video_type"
	VideoModel = "video_model"
	RecType    = "rec_type"
	RecModel   = "rec_model"
	TtsType    = "tts_type"
	TtsModel   = "tts_model"
	RecPhoto   = "rec_photo"
	CronList   = "cron_list"
	CronDel    = "cron_del"
	CronClear  = "cron_clear"
)

var (
	DeepseekLocalModels = map[string]bool{
		LLAVA:                         true,
		deepseek.AzureDeepSeekR1:      true,
		deepseek.OpenRouterDeepSeekR1: true,
		deepseek.OpenRouterDeepSeekR1DistillLlama70B: true,
		deepseek.OpenRouterDeepSeekR1DistillLlama8B:  true,
		deepseek.OpenRouterDeepSeekR1DistillQwen14B:  true,
		deepseek.OpenRouterDeepSeekR1DistillQwen1_5B: true,
		deepseek.OpenRouterDeepSeekR1DistillQwen32B:  true,
	}
	
	GeminiModels = map[string]bool{
		ModelGemini25Pro:       true,
		ModelGemini25Flash:     true,
		ModelGemini20Flash:     true,
		ModelGemini20FlashLite: true,
		ModelGemini15Pro:       true,
		ModelGemini15Flash:     true,
	}
	
	GeminiRecModels = map[string]bool{
		ModelGemini25Pro:   true,
		ModelGemini25Flash: true,
		ModelGemini20Flash: true,
	}
	
	GeminiTTSModels = map[string]bool{
		Gemini2_5FlashPreviewTTS: true,
	}
	
	VolImageModels = map[string]bool{
		DoubaoSeed16VisionPro: true,
	}
	
	VolVideoModels = map[string]bool{
		DoubaoSeedance1_0Pro: true,
	}
	
	VolRecModels = map[string]bool{
		DoubaoSeed16VisionPro: true,
	}
	
	VolTTSModels = map[string]bool{
		VolTTS: true,
		VolIcl: true,
	}
	
	AliyunImageModels = map[string]bool{
		QwenImagePlus: true,
	}
	
	AliyunVideoModels = map[string]bool{
		Wan2_5T2VPreview: true,
	}
	
	AliyunRecModels = map[string]bool{
		QwenVlMax:       true,
		QwenAudioTurbo:  true,
		QwenVlMaxLatest: true,
		QVQMax:          true,
		QVQMaxLatest:    true,
	}
	
	AliyunTTSModels = map[string]bool{
		Qwen3TTSFlash: true,
	}
	
	//OpenAIImageModels = map[string]bool{
	//	ModelImageGPT: true,
	//}
	//
	//OpenAiRecModels = map[string]bool{
	//	openai.Whisper1: true,
	//	ChatGPT4_0:      true,
	//}
	//
	GeminiVideoModels = map[string]bool{
		GeminiVideoVeo3_1Preview:     true,
		GeminiVideoVeo3_1FastPreview: true,
		GeminiVideoFastVeo3:          true,
		GeminiVideoVeo3:              true,
		GeminiVideoVeo2:              true,
	}
	
	GeminiImageModels = map[string]bool{
		Imagen4_0FastGenerate002: true,
		GeminiImageGenV2_5:       true,
		Imagen3_0Generate002:     true,
		Imagen4_0Generate001:     true,
		Gemini2_5FlashPreview:    true,
	}
	
	DeepseekModels = map[string]bool{
		deepseek.DeepSeekChat:     true,
		deepseek.DeepSeekReasoner: true,
		deepseek.DeepSeekCoder:    true,
	}
	
	VolModels = map[string]bool{
		// doubao Seed 1.6
		ModelDoubaoSeed16:         true,
		ModelDoubaoSeed16Flash:    true,
		ModelDoubaoSeed16Thinking: true,
		
		// doubao 1.5 系列
		ModelDoubao15ThinkingPro:     true,
		ModelDoubao15ThinkingProM428: true,
		ModelDoubao15ThinkingProM415: true,
		ModelDoubao15VisionPro428:    true,
		ModelDoubao15VisionPro328:    true,
		
		// DeepSeek R1
		ModelDeepSeekR1_528:    true,
		ModelDeepSeekR1_120:    true,
		ModelDeepSeekR1Qwen32b: true,
		ModelDeepSeekR1Qwen7b:  true,
		
		// doubao-1.5
		ModelDoubao15VisionPro32k: true,
		ModelDoubao15VisionLite:   true,
	}
	
	AliyunModel = map[string]bool{
		qwen.QwenLong:           true,
		qwen.QwenTurbo:          true,
		qwen.QwenPlus:           true,
		qwen.QwenMax:            true,
		qwen.QwenMax1201:        true,
		qwen.QwenMaxLongContext: true,
		
		// multi-modal model.
		qwen.QwenVLPlus:     true,
		qwen.QwenVLMax:      true,
		qwen.QwenAudioTurbo: true,
		
		ModelDeepSeekR1_528: true,
		ModelDeepSeekR1:     true,
		ModelDeepseekV3:     true,
	}
)

type MsgInfo struct {
	MsgId       string
	Content     string
	PartContent string
	SendLen     int
	Finished    bool
}

type ImgResponse struct {
	Code    int              `json:"code"`
	Data    *ImgResponseData `json:"data"`
	Message string           `json:"message"`
	Status  int              `json:"status"`
}

type ImgResponseData struct {
	AlgorithmBaseResp struct {
		StatusCode    int    `json:"status_code"`
		StatusMessage string `json:"status_message"`
	} `json:"algorithm_base_resp"`
	ImageUrls        []string `json:"image_urls"`
	PeResult         string   `json:"pe_result"`
	PredictTagResult string   `json:"predict_tag_result"`
	RephraserResult  string   `json:"rephraser_result"`
}

type LLMConfig struct {
	TxtType    string `json:"txt_type"`
	TxtModel   string `json:"txt_model"`
	ImgType    string `json:"img_type"`
	ImgModel   string `json:"img_model"`
	VideoType  string `json:"video_type"`
	VideoModel string `json:"video_model"`
	RecType    string `json:"rec_type"`
	RecModel   string `json:"rec_model"`
	TTSType    string `json:"tts_type"`
	TTSModel   string `json:"tts_model"`
}

type ContextState struct {
	Token     int
	RecordID  int64
	SkipCheck bool
	UseRecord bool
}
