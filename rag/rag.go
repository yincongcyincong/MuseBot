package rag

import (
	"context"
	"errors"

	"github.com/cohesion-org/deepseek-go"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type DeepSeekLLM struct {
	Client *deepseek.Client

	LLM *llm.LLM
}

func NewDeepSeekLLM(options ...llm.Option) *DeepSeekLLM {
	dp := &DeepSeekLLM{
		Client: deepseek.NewClient(*conf.DeepseekToken),

		LLM: llm.NewLLM(options...),
	}

	for _, o := range options {
		o(dp.LLM)
	}
	return dp
}

func (l *DeepSeekLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l, prompt, options...)
}

func (l *DeepSeekLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	doc, err := conf.Store.SimilaritySearch(ctx, l.LLM.Content, 3)
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
		l.LLM.Content = tmpContent
	}

	err = l.LLM.LLMClient.CallLLMAPI(ctx, l.LLM.Content, l.LLM)
	if err != nil {
		logger.Error("error calling DeepSeek API", "err", err)
		return nil, errors.New("error calling DeepSeek API")
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: l.LLM.WholeContent,
			},
		},
	}

	return resp, nil
}
