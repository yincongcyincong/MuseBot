package param

const (
	DeepSeek = "deepseek"

	ImageTokenUsage = 10000
	VideoTokenUsage = 20000
)

type MsgInfo struct {
	MsgId       int
	Content     string
	SendLen     int
	FullContent string
	Token       int
}
