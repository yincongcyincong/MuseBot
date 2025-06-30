package llm

import (
	"context"
	"errors"
	"time"
	
	godeepseek "github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/revrost/go-openrouter"
	"github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
	"google.golang.org/genai"
)

const (
	OneMsgLen       = 3896
	FirstSendLen    = 30
	NonFirstSendLen = 500
	MostLoop        = 5
)

var (
	ToolsJsonErr = errors.New("tools json error")
)

type LLM struct {
	MessageChan chan *param.MsgInfo
	Update      tgbotapi.Update
	Bot         *tgbotapi.BotAPI
	Content     string // question from user
	Model       string
	Token       int
	
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
	CallLLMAPI(ctx context.Context, prompt string, l *LLM) error
	
	GetMessages(userId int64, prompt string)
	
	Send(ctx context.Context, l *LLM) error
	
	GetUserMessage(msg string)
	
	GetAssistantMessage(msg string)
	
	AppendMessages(client LLMClient)
	
	SyncSend(ctx context.Context, l *LLM) (string, error)
	
	GetModel(l *LLM)
}

func (l *LLM) GetContent() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	chatId, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	defer func() {
		if err := recover(); err != nil {
			logger.Error("GetContent panic err", "err", err)
			utils.SendMsg(chatId, "GetContent panic", l.Bot, msgId, "")
		}
		utils.DecreaseUserChat(l.Update)
		close(l.MessageChan)
	}()
	
	text, err := utils.GetContent(l.Update, l.Bot, l.Content)
	if err != nil {
		logger.Error("get content fail", "err", err)
		utils.SendMsg(chatId, err.Error(), l.Bot, msgId, "")
		return
	}
	l.Content = text
	err = l.LLMClient.CallLLMAPI(ctx, text, l)
	if err != nil {
		logger.Error("Error calling DeepSeek API", "err", err)
		utils.SendMsg(chatId, err.Error(), l.Bot, msgId, "")
	}
}

func NewLLM(opts ...Option) *LLM {
	
	l := new(LLM)
	for _, opt := range opts {
		opt(l)
	}
	
	switch *conf.Type {
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
	case param.OpenRouter:
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

func (l *LLM) sendMsg(msgInfoContent *param.MsgInfo, content string) *param.MsgInfo {
	// exceed max telegram one message length
	if utils.Utf16len(msgInfoContent.Content) > OneMsgLen {
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

func WithUpdate(update tgbotapi.Update) Option {
	return func(p *LLM) {
		p.Update = update
	}
}

func WithBot(bot *tgbotapi.BotAPI) Option {
	return func(p *LLM) {
		p.Bot = bot
	}
}

func WithMessageChan(messageChan chan *param.MsgInfo) Option {
	return func(p *LLM) {
		p.MessageChan = messageChan
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
