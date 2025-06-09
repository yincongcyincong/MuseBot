package rag

import (
	"context"
	"errors"

	"github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	local_deepseek "github.com/yincongcyincong/telegram-deepseek-bot/deepseek"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

const (
	OneMsgLen       = 3896
	FirstSendLen    = 30
	NonFirstSendLen = 500
)

var (
	toolsJsonErr = errors.New("tools json error")
)

type DeepSeekLLM struct {
	Client *deepseek.Client

	DpReq *local_deepseek.DeepseekReq
}

func NewDeepSeekLLM(options ...Option) *DeepSeekLLM {
	dp := &DeepSeekLLM{
		Client: deepseek.NewClient(*conf.DeepseekToken),

		DpReq: &local_deepseek.DeepseekReq{
			ToolCall:           []deepseek.ToolCall{},
			ToolMessage:        []deepseek.ChatCompletionMessage{},
			CurrentToolMessage: []deepseek.ChatCompletionMessage{},
		},
	}

	for _, o := range options {
		o(dp)
	}
	return dp
}

type Option func(p *DeepSeekLLM)

func WithModel(model string) Option {
	return func(p *DeepSeekLLM) {
		p.DpReq.Model = model
	}
}

func WithContent(content string) Option {
	return func(p *DeepSeekLLM) {
		p.DpReq.Content = content
	}
}

func WithUpdate(update tgbotapi.Update) Option {
	return func(p *DeepSeekLLM) {
		p.DpReq.Update = update
	}
}

func WithBot(bot *tgbotapi.BotAPI) Option {
	return func(p *DeepSeekLLM) {
		p.DpReq.Bot = bot
	}
}

func WithMessageChan(messageChan chan *param.MsgInfo) Option {
	return func(p *DeepSeekLLM) {
		p.DpReq.MessageChan = messageChan
	}
}

func (l *DeepSeekLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l, prompt, options...)
}

func (l *DeepSeekLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	doc, err := conf.Store.SimilaritySearch(ctx, l.DpReq.Content, 3)
	if err != nil {
		logger.Error("request vector db fail", "err", err)
	}
	if len(doc) != 0 {
		tmpContent := ""
		for _, msg := range messages {
			for _, part := range msg.Parts {
				tmpContent += part.(llms.TextContent).Text
			}
		}
		l.DpReq.Content = tmpContent
	}

	err = l.DpReq.CallDeepSeekAPI(ctx, l.DpReq.Content)
	if err != nil {
		logger.Error("error calling DeepSeek API", "err", err)
		return nil, errors.New("error calling DeepSeek API")
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: l.DpReq.DeepSeekContent,
			},
		},
	}

	return resp, nil
}
