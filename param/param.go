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
	Aliyun = "aliyun"

	Gemini = "gemini"
	// Gemini 4 系列
	ModelGemini4Carbon string = "gemini-4-carbon"
	ModelGemini4Argon  string = "gemini-4-argon"
	// Gemini 3.x 系列
	ModelGemini38Flash     string = "gemini-3.8-flash"
	ModelGemini36Flash     string = "gemini-3.6-flash"
	ModelGemini35FlashLite string = "gemini-3.5-flash-lite"
	ModelGemini31Pro       string = "gemini-3.1-pro-preview"
	ModelGemini30Flash     string = "gemini-3-flash-preview"
	// 2.5 系列已被官方标记弃用（2026-12-01 起自动路由到新模型），保留兼容存量配置
	ModelGemini25Pro   string = "gemini-2.5-pro"
	ModelGemini25Flash string = "gemini-2.5-flash"

	// Gemini 向量模型（现行；旧 embedding-001 已下线）
	ModelGeminiEmbedding001 string = "gemini-embedding-001"

	OpenAi = "openai"

	LLAVA = "llava:latest"

	DiscordNewMode = "new"

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

	// DeepSeek 官方 API 现行模型（V4 系列；deepseek-chat/reasoner 为旧别名）
	ModelDeepseekFlash = "deepseek-flash"
	ModelDeepseekV4Pro = "deepseek-v4-pro"

	// doubao-1.5
	ModelDoubao15VisionPro32k = "doubao-1.5-vision-pro-32k-250115"
	ModelDoubao15VisionLite   = "doubao-1.5-vision-lite-250315"

	AllRecordType      = -1
	TextRecordType     = 0
	ImageRecordType    = 1
	VideoRecordType    = 2
	WEBRecordType      = 3
	TalkRecordType     = 4
	SummaryRecordType  = 5
	DocumentRecordType = 6

	DefaultContextToken = 128000

	GeminiNanoBanana21     = "gemini-nano-banana-2.1"
	GeminiImage31Flash     = "gemini-3.1-flash-image"
	GeminiImage31FlashLite = "gemini-3.1-flash-lite-image"
	GeminiImage3Pro        = "gemini-3-pro-image"
	GeminiImageGenV2_5     = "gemini-2.5-flash-image"

	GeminiVideoOmni12Flash = "gemini-omni-1.2-flash"
	GeminiVideoOmniFlash   = "gemini-omni-1.1-flash"

	ModelGPT52      = "gpt-5.2"
	ModelGPT52Pro   = "gpt-5.2-pro"
	ModelGPT51      = "gpt-5.1"
	ModelGPT5Mini   = "gpt-5-mini"
	ModelGPT5Nano   = "gpt-5-nano"
	ModelImageGPT15 = "gpt-image-1.5"
	ModelImageGPT   = "gpt-image-1"

	DoubaoSeed16VisionPro = "doubao-seed-1-6-250615"

	QwenImagePlus   = "qwen-image-plus"
	QwenImageEdit   = "qwen-image-edit"
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

	Gemini38FlashTTS         = "gemini-3.8-flash-tts"
	Gemini38FlashLiteTTS     = "gemini-3.8-flash-lite-tts"
	Gemini31FlashTTSPreview  = "gemini-3.1-flash-tts-preview"
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
	Approve    = "y"
	Deny       = "n"
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
		// Gemini 4
		ModelGemini4Carbon: true,
		ModelGemini4Argon:  true,
		// Gemini 3.x
		ModelGemini38Flash:     true,
		ModelGemini36Flash:     true,
		ModelGemini35FlashLite: true,
		ModelGemini31Pro:       true,
		ModelGemini30Flash:     true,
		// 2.5 legacy（2026-12-01 起自动路由）
		ModelGemini25Pro:   true,
		ModelGemini25Flash: true,
	}

	GeminiRecModels = map[string]bool{
		ModelGemini4Carbon:     true,
		ModelGemini4Argon:      true,
		ModelGemini38Flash:     true,
		ModelGemini36Flash:     true,
		ModelGemini35FlashLite: true,
		// 2.5 legacy
		ModelGemini25Pro:   true,
		ModelGemini25Flash: true,
	}

	GeminiTTSModels = map[string]bool{
		Gemini38FlashTTS:         true,
		Gemini38FlashLiteTTS:     true,
		Gemini31FlashTTSPreview:  true,
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
		QwenImageEdit: true,
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
		GeminiVideoOmni12Flash: true,
		GeminiVideoOmniFlash:   true,
	}

	GeminiImageModels = map[string]bool{
		GeminiNanoBanana21:     true,
		GeminiImage3Pro:        true,
		GeminiImage31Flash:     true,
		GeminiImage31FlashLite: true,
		GeminiImageGenV2_5:     true,
	}

	DeepseekModels = map[string]bool{
		// V4 现行模型
		ModelDeepseekFlash: true,
		ModelDeepseekV4Pro: true,
		// 旧别名，仍被 API 接受
		deepseek.DeepSeekChat:     true,
		deepseek.DeepSeekReasoner: true,
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

// DocInfo carries a chat document received from an IM platform.
type DocInfo struct {
	FileName string
	Content  []byte
}

type MCPResp struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}
