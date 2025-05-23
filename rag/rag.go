package rag

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
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
	Client      *deepseek.Client
	MessageChan chan *param.MsgInfo
	Update      tgbotapi.Update
	Bot         *tgbotapi.BotAPI
	Content     string
	Model       string
	Token       int

	ToolCall           []deepseek.ToolCall
	DeepSeekContent    string
	ToolMessage        []deepseek.ChatCompletionMessage
	CurrentToolMessage []deepseek.ChatCompletionMessage
}

func NewDeepSeekLLM(options ...Option) *DeepSeekLLM {
	dp := &DeepSeekLLM{
		Client: deepseek.NewClient(*conf.DeepseekToken),

		ToolCall:           []deepseek.ToolCall{},
		ToolMessage:        []deepseek.ChatCompletionMessage{},
		CurrentToolMessage: []deepseek.ChatCompletionMessage{},
	}

	for _, o := range options {
		o(dp)
	}
	return dp
}

type Option func(p *DeepSeekLLM)

func WithModel(model string) Option {
	return func(p *DeepSeekLLM) {
		p.Model = model
	}
}

func WithContent(content string) Option {
	return func(p *DeepSeekLLM) {
		p.Content = content
	}
}

func WithUpdate(update tgbotapi.Update) Option {
	return func(p *DeepSeekLLM) {
		p.Update = update
	}
}

func WithBot(bot *tgbotapi.BotAPI) Option {
	return func(p *DeepSeekLLM) {
		p.Bot = bot
	}
}

func WithMessageChan(messageChan chan *param.MsgInfo) Option {
	return func(p *DeepSeekLLM) {
		p.MessageChan = messageChan
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

	msg0 := messages[0]
	part := msg0.Parts[0]
	l.Content = part.(llms.TextContent).Text

	err := l.callDeepSeekAPI(ctx, l.Content)
	if err != nil {
		logger.Error("error calling DeepSeek API", "err", err)
		return nil, errors.New("error calling DeepSeek API")
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: l.DeepSeekContent,
			},
		},
	}

	return resp, nil
}

// callDeepSeekAPI request DeepSeek API and get response
func (l *DeepSeekLLM) callDeepSeekAPI(ctx context.Context, prompt string) error {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	l.Model = deepseek.DeepSeekChat
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}

	messages := make([]deepseek.ChatCompletionMessage, 0)

	msgRecords := db.GetMsgRecord(userId)
	if msgRecords != nil {
		aqs := msgRecords.AQs
		if len(aqs) > 10 {
			aqs = aqs[len(aqs)-10:]
		}

		for i, record := range aqs {
			if record.Answer != "" && record.Question != "" {
				logger.Info("context content", "dialog", i, "question:", record.Question,
					"toolContent", record.Content, "answer:", record.Answer)
				messages = append(messages, deepseek.ChatCompletionMessage{
					Role:    constants.ChatMessageRoleUser,
					Content: record.Question,
				})
				if record.Content != "" {
					toolsMsgs := make([]deepseek.ChatCompletionMessage, 0)
					err = json.Unmarshal([]byte(record.Content), &toolsMsgs)
					if err != nil {
						logger.Error("Error unmarshalling tools json", "err", err)
					} else {
						messages = append(messages, toolsMsgs...)
					}
				}
				messages = append(messages, deepseek.ChatCompletionMessage{
					Role:    constants.ChatMessageRoleAssistant,
					Content: record.Answer,
				})
			}
		}
	}

	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: prompt,
	})

	logger.Info("msg receive", "userID", userId, "prompt", prompt)

	return l.send(ctx, messages)
}

func (l *DeepSeekLLM) send(ctx context.Context, messages []deepseek.ChatCompletionMessage) error {
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	// set deepseek proxy
	httpClient := &http.Client{
		Timeout: 30 * time.Minute,
	}

	if *conf.DeepseekProxy != "" {
		proxy, err := url.Parse(*conf.DeepseekProxy)
		if err != nil {
			logger.Error("parse deepseek proxy error", "err", err)
		} else {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	client, err := deepseek.NewClientWithOptions(*conf.DeepseekToken,
		deepseek.WithBaseURL(*conf.CustomUrl), deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.Error("Error creating deepseek client", "err", err)
		return err
	}

	request := &deepseek.StreamChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: deepseek.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
	}

	request.Messages = messages

	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", updateMsgID, "err", err)
		return err
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}

	hasTools := false
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logger.Info("Stream finished", "updateMsgID", updateMsgID)
			break
		}
		if err != nil {
			logger.Warn("Stream error", "updateMsgID", updateMsgID, "err", err)
			break
		}
		for _, choice := range response.Choices {
			if len(choice.Delta.ToolCalls) > 0 {
				hasTools = true
				err = l.requestToolsCall(ctx, choice)
				if err != nil {
					if errors.Is(err, toolsJsonErr) {
						continue
					} else {
						logger.Error("requestToolsCall error", "updateMsgID", updateMsgID, "err", err)
					}
				}
			}

			if !hasTools {
				l.sendMsg(msgInfoContent, choice)
			}
		}

		if response.Usage != nil {
			l.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(l.Token))
		}
	}

	if !hasTools || len(l.CurrentToolMessage) == 0 {
		l.MessageChan <- msgInfoContent

		data, _ := json.Marshal(l.ToolMessage)
		db.InsertMsgRecord(userId, &db.AQ{
			Question: l.Content,
			Answer:   l.DeepSeekContent,
			Content:  string(data),
			Token:    l.Token,
		}, true)
	} else {
		l.CurrentToolMessage = append([]deepseek.ChatCompletionMessage{
			{
				Role:      deepseek.ChatMessageRoleAssistant,
				Content:   l.DeepSeekContent,
				ToolCalls: l.ToolCall,
			},
		}, l.CurrentToolMessage...)

		l.ToolMessage = append(l.ToolMessage, l.CurrentToolMessage...)
		messages = append(messages, l.CurrentToolMessage...)
		l.CurrentToolMessage = make([]deepseek.ChatCompletionMessage, 0)
		l.ToolCall = make([]deepseek.ToolCall, 0)
		return l.send(ctx, messages)
	}

	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
	return nil
}

func (l *DeepSeekLLM) sendMsg(msgInfoContent *param.MsgInfo, choice deepseek.StreamChoices) {
	// exceed max telegram one message length
	if utils.Utf16len(msgInfoContent.Content) > OneMsgLen {
		l.MessageChan <- msgInfoContent
		msgInfoContent = &param.MsgInfo{
			SendLen: NonFirstSendLen,
		}
	}

	msgInfoContent.Content += choice.Delta.Content
	l.DeepSeekContent += choice.Delta.Content
	if len(msgInfoContent.Content) > msgInfoContent.SendLen {
		l.MessageChan <- msgInfoContent
		msgInfoContent.SendLen += NonFirstSendLen
	}
}

func (l *DeepSeekLLM) requestToolsCall(ctx context.Context, choice deepseek.StreamChoices) error {

	for _, toolCall := range choice.Delta.ToolCalls {
		property := make(map[string]interface{})

		if toolCall.Function.Name != "" {
			l.ToolCall = append(l.ToolCall, toolCall)
			l.ToolCall[len(l.ToolCall)-1].Function.Name = toolCall.Function.Name
		}

		if toolCall.ID != "" {
			l.ToolCall[len(l.ToolCall)-1].ID = toolCall.ID
		}

		if toolCall.Type != "" {
			l.ToolCall[len(l.ToolCall)-1].Type = toolCall.Type
		}

		if toolCall.Function.Arguments != "" {
			l.ToolCall[len(l.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}

		err := json.Unmarshal([]byte(l.ToolCall[len(l.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return toolsJsonErr
		}

		mc, err := clients.GetMCPClientByToolName(l.ToolCall[len(l.ToolCall)-1].Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return err
		}

		toolsData, err := mc.ExecTools(ctx, l.ToolCall[len(l.ToolCall)-1].Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return err
		}
		l.CurrentToolMessage = append(l.CurrentToolMessage, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: l.ToolCall[len(l.ToolCall)-1].ID,
		})

	}

	logger.Info("send tool request", "function", l.ToolCall[len(l.ToolCall)-1].Function.Name,
		"toolCall", l.ToolCall[len(l.ToolCall)-1].ID, "argument", l.ToolCall[len(l.ToolCall)-1].Function.Arguments)

	return nil

}
