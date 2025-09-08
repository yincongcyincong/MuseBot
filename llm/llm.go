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
	RecordId    int64
	
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
	
	l.GetMessages(l.UserId, l.GetContent(l.Content))
	
	logger.Info("msg receive", "userID", l.UserId, "prompt", l.Content)
	
	l.LLMClient.GetModel(l)
	
	err := l.LLMClient.Send(ctx, l)
	if err != nil {
		logger.Error("Error calling LLM API", "err", err)
		return err
	}
	
	err = l.InsertOrUpdate()
	if err != nil {
		logger.Error("insert or update record", "err", err)
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
	case param.Ollama:
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

func (l *LLM) InsertOrUpdate() error {
	if l.RecordId == 0 {
		db.InsertMsgRecord(l.UserId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Token:    l.Token,
			Mode:     *conf.BaseConfInfo.Type,
		}, true)
		return nil
	}
	
	db.InsertMsgRecord(l.UserId, &db.AQ{
		Question: l.Content,
		Answer:   l.WholeContent,
	}, false)
	err := db.UpdateRecordInfo(&db.Record{
		ID:     l.RecordId,
		Answer: l.WholeContent,
		Token:  l.Token,
		Mode:   *conf.BaseConfInfo.Type,
	})
	if err != nil {
		logger.Error("update record fail", "err", err)
		return err
	}
	
	return nil
}

func (l *LLM) GetMessages(userId string, prompt string) {
	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		aqs := msgRecords.AQs
		if len(aqs) > *conf.BaseConfInfo.MaxQAPair {
			aqs = aqs[len(aqs)-*conf.BaseConfInfo.MaxQAPair:]
		}
		
		for i, record := range aqs {
			
			logger.Info("context content", "dialog", i, "question:", record.Question,
				"toolContent", record.Content, "answer:", record.Answer)
			if record.Question != "" {
				l.LLMClient.GetUserMessage(record.Question)
			}
			
			if record.Answer != "" {
				l.LLMClient.GetAssistantMessage(record.Question)
			}
		}
	}
	
	if *conf.BaseConfInfo.Type != "gemini" {
		l.LLMClient.GetUserMessage(prompt)
	}
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

func WithRecordId(recordId int64) Option {
	return func(p *LLM) {
		p.RecordId = recordId
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
