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
	// 火山方舟现行模型。旧的 seed-1.6 / doubao-1.5 / deepseek-r1 快照版已按
	// 第 6–10 批下线公告于 2026-03 至 2026-11 间全部 EOS，官方迁移目标如下：
	// https://ark.volcengine.com/region:cn-beijing/docs/model-deprecation-notice
	DoubaoSeed20Lite  = "doubao-seed-2-0-lite-260428"
	DoubaoSeed20Mini  = "doubao-seed-2-0-mini-260428"
	DoubaoSeed21Lite  = "doubao-seed-2-1-lite-260915"
	DeepseekV41Flash  = "deepseek-v4-1-flash-260910"
	DeepseekV4FlashGA = "deepseek-v4-flash-ga-260731"

	DoubaoSeedance20     = "doubao-seedance-2-0-260128"
	DoubaoSeedance20Fast = "doubao-seedance-2-0-fast-260128"
	DoubaoSeedance20Mini = "doubao-seedance-2-0-mini-260615"

	ModelDeepSeekR1 = "deepseek-r1"

	ModelDeepseekV3         = "deepseek-v3.1"
	ModelDeepSeekR1Qwen32bQ = "deepseek-r1-distill-qwen-32b"
	ModelDeepSeekR1Qwen7bQ  = "deepseek-r1-distill-qwen-7b"

	// DeepSeek 官方 API 现行模型（V4 系列；deepseek-chat/reasoner 为旧别名）
	ModelDeepseekFlash = "deepseek-flash"
	ModelDeepseekV4Pro = "deepseek-v4-pro"

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

	GeminiVideoOmni12Flash = "gemini-omni-1.2-flash"

	ModelGPT52      = "gpt-5.2"
	ModelGPT52Pro   = "gpt-5.2-pro"
	ModelGPT51      = "gpt-5.1"
	ModelGPT5Mini   = "gpt-5-mini"
	ModelGPT5Nano   = "gpt-5-nano"
	ModelGPT6Luna   = "gpt-6-luna"

	ModelImageGPT26Sunburst = "gpt-image-2.6-sunburst"
	ModelImageGPT26Flare    = "gpt-image-2.6-flare"

	// OpenAI speech/transcription models: tts-1 retires on 2027-01-06 and
	// whisper-1 on 2027-02-26 (https://developers.openai.com/api/docs/deprecations)
	ModelGPTRealtime21Mini = "gpt-realtime-2.1-mini"
	ModelGPTTranscribe     = "gpt-transcribe"

	QwenImagePlus   = "qwen-image-plus"
	QwenImageEdit   = "qwen-image-edit"
	QwenVlMaxLatest = "qwen-vl-max-latest"
	QwenVlMax       = "qwen-vl-max"
	QVQMax          = "qvq-max"
	QVQMaxLatest    = "qvq-max-latest"

	QwenAudioTurbo = "qwen-audio-turbo-latest"

	Wan2_5T2VPreview = "wan2.5-t2v-preview"

	VolTTS = "volcano_tts"
	VolIcl = "volcano_icl"

	Qwen3TTSFlash = "qwen3-tts-flash"

	Gemini38FlashTTS     = "gemini-3.8-flash-tts"
	Gemini38FlashLiteTTS = "gemini-3.8-flash-lite-tts"

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
		Gemini38FlashTTS:     true,
		Gemini38FlashLiteTTS: true,
	}

	// Vol whitelists hold only current Ark models; retired snapshots fall back
	// to the defaults in utils.GetUsing*Model (batch 6-10 deprecations)
	VolImageModels = map[string]bool{
		DoubaoSeed20Lite: true,
	}

	VolVideoModels = map[string]bool{
		DoubaoSeedance20Mini: true,
		DoubaoSeedance20:     true,
		DoubaoSeedance20Fast: true,
	}

	VolRecModels = map[string]bool{
		DoubaoSeed20Lite: true,
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
	//	ModelImageGPT26Sunburst: true,
	//	ModelImageGPT26Flare:    true,
	//}
	//
	//OpenAiRecModels = map[string]bool{
	//	ModelGPTTranscribe: true,
	//}
	//

	// OpenAI shut gpt-image-1 down on 2026-10-23 and retires gpt-image-1.5
	// on 2026-12-01. Map the retired names to the recommended replacement so
	// user configs stored before the shutdown keep working.
	// https://developers.openai.com/api/docs/deprecations
	DeprecatedOpenAIImageModels = map[string]string{
		"gpt-image-1":   ModelImageGPT26Sunburst,
		"gpt-image-1.5": ModelImageGPT26Sunburst,
	}

	// Retired or retiring OpenAI speech models (tts-1 family shuts down on
	// 2027-01-06) map to the recommended replacement gpt-realtime-2.1-mini.
	DeprecatedOpenAITTSModels = map[string]string{
		"tts-1":           ModelGPTRealtime21Mini,
		"tts-1-hd":        ModelGPTRealtime21Mini,
		"gpt-4o-mini-tts": ModelGPTRealtime21Mini,
	}

	// Former OpenAI/ChatAnyWhere default text models map to their
	// replacement (gpt-4o family retires on 2027-01-14; chatgpt-4o-latest
	// already shut down on 2026-02-17). Custom relay model names not listed
	// here are passed through untouched.
	DeprecatedOpenAITxtModels = map[string]string{
		"gpt-4o-mini":       ModelGPT6Luna,
		"gpt-4o":            ModelGPT6Luna,
		"chatgpt-4o-latest": ModelGPT6Luna,
	}
	// gemini-omni-1.1-flash shut down on 2026-12-03; stored configs fall back
	// to gemini-omni-1.2-flash. https://ai.google.dev/gemini-api/docs/deprecations
	GeminiVideoModels = map[string]bool{
		GeminiVideoOmni12Flash: true,
	}

	GeminiImageModels = map[string]bool{
		GeminiNanoBanana21:     true,
		GeminiImage3Pro:        true,
		GeminiImage31Flash:     true,
		GeminiImage31FlashLite: true,
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
		// doubao Seed 2.x
		DoubaoSeed21Lite: true,
		DoubaoSeed20Lite: true,
		DoubaoSeed20Mini: true,

		// 方舟托管的 DeepSeek
		DeepseekV41Flash:  true,
		DeepseekV4FlashGA: true,
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

		// Aliyun-hosted DeepSeek entries kept as literals; their availability
		// is tracked by DashScope, not the Volcano Ark deprecation batches
		"deepseek-r1-250528": true,
		ModelDeepSeekR1:      true,
		ModelDeepseekV3:      true,
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
