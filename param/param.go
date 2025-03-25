package param

const (
	DeepSeek = "deepseek"
)

type MsgInfo struct {
	MsgId       int
	Content     string
	SendLen     int
	FullContent string
	Token       int
}
