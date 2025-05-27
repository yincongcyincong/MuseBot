package param

const (
	DeepSeek      = "deepseek"
	DeepSeekLlava = "deepseek-ollama"

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

	LLAVA = "llava:latest"

	ImageTokenUsage = 10000
	VideoTokenUsage = 20000
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
	Status  string           `json:"status"`
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
