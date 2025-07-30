package param

import (
	"github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
)

const (
	DeepSeek      = "deepseek"
	DeepSeekLlava = "deepseek-ollama"
	
	Vol = "vol"
	
	Gemini                        = "gemini"
	ModelGemini25Pro       string = "gemini-2.5-pro"
	ModelGemini25Flash     string = "gemini-2.5-flash"
	ModelGemini20Flash     string = "gemini-2.0-flash"
	ModelGemini20FlashLite string = "gemini-2.0-flash-lite"
	ModelGemini15Pro       string = "gemini-1.5-pro"
	ModelGemini15Flash     string = "gemini-1.5-flash"
	ModelGemini10Ultra     string = "gemini-1.0-ultra"
	ModelGemini10Pro       string = "gemini-1.0-pro"
	ModelGemini10Nano      string = "gemini-1.0-nano"
	
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
	ModelDeepSeekR1_528    = "deepseek-r1-250528"
	ModelDeepSeekR1_120    = "deepseek-r1-250120"
	ModelDeepSeekR1Qwen32b = "deepseek-r1-distill-qwen-32b-250120"
	ModelDeepSeekR1Qwen7b  = "deepseek-r1-distill-qwen-7b-250120"
	
	// doubao-1.5
	ModelDoubao15VisionPro32k = "doubao-1.5-vision-pro-32k-250115"
	ModelDoubao15VisionLite   = "doubao-1.5-vision-lite-250315"
	
	TextRecordType  = 0
	ImageRecordType = 1
	VideoRecordType = 2
	WEBRecordType   = 3
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
		ModelGemini10Ultra:     true,
		ModelGemini10Pro:       true,
		ModelGemini10Nano:      true,
	}
	
	DeepseekModels = map[string]bool{
		deepseek.DeepSeekChat:     true,
		deepseek.DeepSeekReasoner: true,
		deepseek.DeepSeekCoder:    true,
	}
	
	OpenAIModels = map[string]bool{
		openai.GPT3Dot5Turbo0125:       true,
		openai.O1Mini:                  true,
		openai.O1Mini20240912:          true,
		openai.O1Preview:               true,
		openai.O1Preview20240912:       true,
		openai.O1:                      true,
		openai.O120241217:              true,
		openai.O3Mini:                  true,
		openai.O3Mini20250131:          true,
		openai.GPT432K0613:             true,
		openai.GPT432K0314:             true,
		openai.GPT432K:                 true,
		openai.GPT40613:                true,
		openai.GPT40314:                true,
		openai.GPT4o:                   true,
		openai.GPT4o20240513:           true,
		openai.GPT4o20240806:           true,
		openai.GPT4o20241120:           true,
		openai.GPT4oLatest:             true,
		openai.GPT4oMini:               true,
		openai.GPT4oMini20240718:       true,
		openai.GPT4Turbo:               true,
		openai.GPT4Turbo20240409:       true,
		openai.GPT4Turbo0125:           true,
		openai.GPT4Turbo1106:           true,
		openai.GPT4TurboPreview:        true,
		openai.GPT4VisionPreview:       true,
		openai.GPT4:                    true,
		openai.GPT4Dot5Preview:         true,
		openai.GPT4Dot5Preview20250227: true,
		openai.GPT3Dot5Turbo1106:       true,
		openai.GPT3Dot5Turbo0613:       true,
		openai.GPT3Dot5Turbo0301:       true,
		openai.GPT3Dot5Turbo16K:        true,
		openai.GPT3Dot5Turbo16K0613:    true,
		openai.GPT3Dot5Turbo:           true,
		openai.GPT3Dot5TurboInstruct:   true,
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
)

type MsgInfo struct {
	MsgId   int
	Content string
	SendLen int
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
