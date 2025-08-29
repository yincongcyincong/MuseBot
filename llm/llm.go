package llm

import (
	"context"
	"errors"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	"github.com/revrost/go-openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"google.golang.org/genai"
)

const (
	OneMsgLen       = 3896
	FirstSendLen    = 30
	NonFirstSendLen = 500
	MostLoop        = 15
)

var (
	ToolsJsonErr = errors.New("tools json error")
)

type LLM struct {
	MessageChan chan *param.MsgInfo
	HTTPMsgChan chan string
	Content     string // question from user
	Model       string
	Token       int
	
	ChatId    string
	UserId    string
	MsgId     string
	PerMsgLen int
	
	LLMClient LLMClient
	
	DeepseekTools   []godeepseek.Tool
	VolTools        []*model.Tool
	OpenAITools     []openai.Tool
	GeminiTools     []*genai.Tool
	OpenRouterTools []openrouter.Tool
	
	WholeContent string // whole answer from llm
	LoopNum      int
}

type LLMClient interface {
	GetMessages(userId string, prompt string)
	
	Send(ctx context.Context, l *LLM) error
	
	GetUserMessage(msg string)
	
	GetAssistantMessage(msg string)
	
	AppendMessages(client LLMClient)
	
	SyncSend(ctx context.Context, l *LLM) (string, error)
	
	GetModel(l *LLM)
}

func (l *LLM) CallLLM() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	l.LLMClient.GetMessages(l.UserId, l.GetContent(l.Content))
	
	logger.Info("msg receive", "userID", l.UserId, "prompt", l.Content)
	
	l.LLMClient.GetModel(l)
	
	err := l.LLMClient.Send(ctx, l)
	if err != nil {
		logger.Error("Error calling LLM API", "err", err)
		return err
	}
	
	return nil
}

func (l *LLM) GetContent(content string) string {
	if *conf.BaseConfInfo.Character != "" {
		content = *conf.BaseConfInfo.Character + "\n\n" + content
	}
	
	return content
}

func NewLLM(opts ...Option) *LLM {
	
	l := new(LLM)
	for _, opt := range opts {
		opt(l)
	}
	
	switch *conf.BaseConfInfo.Type {
	case param.DeepSeek:
		l.LLMClient = &DeepseekReq{
			ToolCall:           []godeepseek.ToolCall{},
			ToolMessage:        []godeepseek.ChatCompletionMessage{},
			CurrentToolMessage: []godeepseek.ChatCompletionMessage{},
		}
	case param.DeepSeekLlava:
		l.LLMClient = &OllamaDeepseekReq{
			ToolCall:           []godeepseek.ToolCall{},
			ToolMessage:        []godeepseek.ChatCompletionMessage{},
			CurrentToolMessage: []godeepseek.ChatCompletionMessage{},
		}
	case param.Gemini:
		l.LLMClient = &GeminiReq{
			ToolCall:           []*genai.FunctionCall{},
			ToolMessage:        []*genai.Content{},
			CurrentToolMessage: []*genai.Content{},
		}
	case param.OpenAi:
		l.LLMClient = &OpenAIReq{
			ToolCall:           []openai.ToolCall{},
			ToolMessage:        []openai.ChatCompletionMessage{},
			CurrentToolMessage: []openai.ChatCompletionMessage{},
		}
	case param.OpenRouter, param.AI302:
		l.LLMClient = &AIRouterReq{
			ToolCall:           []openrouter.ToolCall{},
			ToolMessage:        []openrouter.ChatCompletionMessage{},
			CurrentToolMessage: []openrouter.ChatCompletionMessage{},
		}
	case param.Vol:
		l.LLMClient = &VolReq{
			ToolCall:           []*model.ToolCall{},
			ToolMessage:        []*model.ChatCompletionMessage{},
			CurrentToolMessage: []*model.ChatCompletionMessage{},
		}
	}
	
	return l
}

func (l *LLM) DirectSendMsg(content string) {
	if len([]byte(content)) > l.PerMsgLen {
		content = string([]byte(content)[:l.PerMsgLen])
	}
	
	if l.MessageChan != nil {
		l.MessageChan <- &param.MsgInfo{
			Content:  content,
			Finished: true,
		}
	}
	
	if l.HTTPMsgChan != nil {
		l.HTTPMsgChan <- content
	}
}

func (l *LLM) SendMsg(msgInfoContent *param.MsgInfo, content string) *param.MsgInfo {
	if l.MessageChan != nil {
		if l.PerMsgLen == 0 {
			l.PerMsgLen = OneMsgLen
		}
		
		// exceed max one message length
		if len([]byte(msgInfoContent.Content)) > l.PerMsgLen {
			msgInfoContent.Finished = true
			l.MessageChan <- msgInfoContent
			msgInfoContent = &param.MsgInfo{
				SendLen: NonFirstSendLen,
			}
		}
		
		msgInfoContent.Content += content
		l.WholeContent += content
		if len(msgInfoContent.Content) > msgInfoContent.SendLen {
			l.MessageChan <- msgInfoContent
			msgInfoContent.SendLen += NonFirstSendLen
		}
		
		return msgInfoContent
	} else {
		l.WholeContent += content
		l.HTTPMsgChan <- content
		return nil
	}
}

func (l *LLM) OverLoop() bool {
	if l.LoopNum >= MostLoop {
		return true
	}
	l.LoopNum++
	return false
}

type Option func(p *LLM)

func WithModel(model string) Option {
	return func(p *LLM) {
		p.Model = model
	}
}

func WithContent(content string) Option {
	return func(p *LLM) {
		p.Content = content
	}
}

func WithPerMsgLen(perMsgLen int) Option {
	return func(p *LLM) {
		p.PerMsgLen = perMsgLen
	}
}

func WithMessageChan(messageChan chan *param.MsgInfo) Option {
	return func(p *LLM) {
		p.MessageChan = messageChan
	}
}

func WithHTTPMsgChan(messageChan chan string) Option {
	return func(p *LLM) {
		p.HTTPMsgChan = messageChan
	}
}

func WithChatId(chatId string) Option {
	return func(p *LLM) {
		p.ChatId = chatId
	}
}

func WithUserId(userId string) Option {
	return func(p *LLM) {
		p.UserId = userId
	}
}

func WithMsgId(msgId string) Option {
	return func(p *LLM) {
		p.MsgId = msgId
	}
}

func WithToken(token int) Option {
	return func(p *LLM) {
		p.Token = token
	}
}

func WithTaskTools(taskTool *conf.AgentInfo) Option {
	return func(p *LLM) {
		if taskTool == nil {
			p.DeepseekTools = nil
			p.VolTools = nil
			p.OpenAITools = nil
			p.GeminiTools = nil
			p.OpenRouterTools = nil
			return
		}
		p.DeepseekTools = taskTool.DeepseekTool
		p.VolTools = taskTool.VolTool
		p.OpenAITools = taskTool.OpenAITools
		p.GeminiTools = taskTool.GeminiTools
		p.OpenRouterTools = taskTool.OpenRouterTools
	}
}

func estimateTokens(text string) int {
	count := 0
	for _, r := range text {
		if r <= 127 {
			count += 1
		} else {
			count += 1
		}
	}
	englishApprox := count / 4
	if englishApprox == 0 {
		englishApprox = 1
	}
	return englishApprox
}

func TruncateAQsByToken(aqs []*db.AQ, maxToken int) []*db.AQ {
	if len(aqs) == 0 {
		return []*db.AQ{}
	}
	
	totalToken := 0
	var truncated []*db.AQ
	
	// 从最新消息开始向前遍历
	for i := len(aqs) - 1; i >= 0; i-- {
		a := aqs[i]
		token := estimateTokens(a.Question) + estimateTokens(a.Answer)
		if totalToken+token > maxToken {
			break
		}
		totalToken += token
		truncated = append([]*db.AQ{a}, truncated...)
	}
	
	return truncated
}
