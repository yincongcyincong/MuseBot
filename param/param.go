package param

const (
	DeepSeek      = "deepseek"
	DeepSeekLlava = "deepseek-ollama"

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
