package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	deepseekUtils "github.com/cohesion-org/deepseek-go/utils"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type OllamaReq struct {
	Tools      LLMToolState
	OllamaMsgs []deepseek.ChatCompletionMessage
}

func (o OllamaReq) GetModel(l *LLM) {
	userInfo := db.GetCtxUserInfo(l.Ctx)
	model := ""
	if userInfo != nil && userInfo.LLMConfigRaw != nil {
		model = userInfo.LLMConfigRaw.TxtModel
	}
	switch utils.GetTxtType(db.GetCtxUserInfo(l.Ctx).LLMConfigRaw) {
	case param.Ollama:
		l.Model = "deepseek-r1"
		if userInfo != nil && model != "" {
			logger.InfoCtx(l.Ctx, "User info", "userID", userInfo.UserId, "mode", model)
			l.Model = model
		}
	}
}

func (o *OllamaReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	start := time.Now()

	// set deepseek proxy
	client := GetDeepseekClient(ctx)
	request := &deepseek.StreamChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: deepseek.StreamOptions{
			IncludeUsage: true,
		},
		Tools: l.DeepseekTools,
	}

	currentLLMSamplingOptions().applyToDeepseekStream(request)

	request.Messages = o.OllamaMsgs

	stream, err := retryLLMRequest(l, func() (*Stream, error) {
		return requestDeepseek(ctx, client, request)
	})
	if err != nil || stream == nil {
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return fmt.Errorf("request fail %v %v", err, stream)
	}
	defer stream.Close()

	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())

	msgInfoContent, hasTools, err := receiveLLMStream(ctx, l, func() (LLMStreamChunk, error) {
		response, err := Receive(stream)
		if err != nil {
			return LLMStreamChunk{}, err
		}
		return normalizeOllamaStreamChunk(response), nil
	}, &o.Tools)
	if err != nil {
		return err
	}

	sendLLMStreamOutput(l, msgInfoContent, hasTools)
	if hasTools && len(o.Tools.Results) != 0 {
		assistantMessage := deepseek.ChatCompletionMessage{
			Role:      deepseek.ChatMessageRoleAssistant,
			Content:   l.WholeContent,
			ToolCalls: deepseekToolCalls(o.Tools.Calls),
		}
		o.OllamaMsgs = append(o.OllamaMsgs, assistantMessage)
		o.OllamaMsgs = append(o.OllamaMsgs, deepseekToolMessages(o.Tools.Results)...)
		o.Tools.reset()
		return o.Send(ctx, l)
	}

	return nil
}

func deepseekToolCalls(calls []LLMToolCall) []deepseek.ToolCall {
	toolCalls := make([]deepseek.ToolCall, 0, len(calls))
	for _, call := range calls {
		toolCalls = append(toolCalls, deepseek.ToolCall{
			ID:   call.ID,
			Type: call.Type,
			Function: deepseek.ToolCallFunction{
				Name:      call.Name,
				Arguments: call.Arguments,
			},
		})
	}
	return toolCalls
}

func deepseekStreamToolCalls(calls []deepseek.ToolCall) []LLMToolCall {
	toolCalls := make([]LLMToolCall, 0, len(calls))
	for _, call := range calls {
		toolCalls = append(toolCalls, LLMToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Type:      call.Type,
			Arguments: call.Function.Arguments,
		})
	}
	return toolCalls
}

func deepseekToolMessages(results []LLMToolResult) []deepseek.ChatCompletionMessage {
	messages := make([]deepseek.ChatCompletionMessage, 0, len(results))
	for _, result := range results {
		messages = append(messages, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    result.Content,
			ToolCallID: result.ID,
		})
	}
	return messages
}

func normalizeOllamaStreamChunk(response *deepseek.StreamChatCompletionResponse) LLMStreamChunk {
	chunk := LLMStreamChunk{}
	for _, choice := range response.Choices {
		chunk.Content += choice.Delta.Content
		chunk.ToolCalls = append(chunk.ToolCalls, deepseekStreamToolCalls(choice.Delta.ToolCalls)...)
	}
	if response.Usage != nil {
		chunk.Usage = response.Usage.TotalTokens
		chunk.HasUsage = true
	}
	return chunk
}

func (o *OllamaReq) GetUserMessage(msg string) {
	o.GetMessage(constants.ChatMessageRoleUser, msg)
}

func (o *OllamaReq) GetAssistantMessage(msg string) {
	o.GetMessage(constants.ChatMessageRoleAssistant, msg)
}

func (o *OllamaReq) GetSystemMessage(msg string) {
	o.GetMessage(constants.ChatMessageRoleSystem, msg)
}

func (o *OllamaReq) GetImageMessage(image [][]byte, msg string) {}

func (o *OllamaReq) GetAudioMessage(audio []byte, msg string) {}

func (o *OllamaReq) AppendMessages(client LLMClient) {
	if len(o.OllamaMsgs) == 0 {
		o.OllamaMsgs = make([]deepseek.ChatCompletionMessage, 0)
	}

	o.OllamaMsgs = append(o.OllamaMsgs, client.(*OllamaReq).OllamaMsgs...)
}

func (o *OllamaReq) GetMessage(role, msg string) {
	if len(o.OllamaMsgs) == 0 {
		o.OllamaMsgs = []deepseek.ChatCompletionMessage{
			{
				Role:    role,
				Content: msg,
			},
		}
		return
	}

	o.OllamaMsgs = append(o.OllamaMsgs, deepseek.ChatCompletionMessage{
		Role:    role,
		Content: msg,
	})
}

func (o *OllamaReq) SyncSend(ctx context.Context, l *LLM) (string, error) {

	start := time.Now()

	client := GetDeepseekClient(ctx)
	request := &deepseek.ChatCompletionRequest{
		Model:    l.Model,
		Messages: o.OllamaMsgs,
		Tools:    l.DeepseekTools,
	}

	currentLLMSamplingOptions().applyToDeepseek(request)

	// assign task
	response, err := retryLLMRequest(l, func() (*deepseek.ChatCompletionResponse, error) {
		return client.CreateChatCompletion(ctx, request)
	})
	if err != nil || response == nil {
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return "", fmt.Errorf("request fail %v %v", err, response)
	}

	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())

	if len(response.Choices) == 0 {
		logger.ErrorCtx(l.Ctx, "response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}

	l.Cs.Token += response.Usage.TotalTokens
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		toolCalls := deepseekStreamToolCalls(response.Choices[0].Message.ToolCalls)
		o.GetAssistantMessage("")
		o.OllamaMsgs[len(o.OllamaMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		if err := o.Tools.observeComplete(ctx, l, toolCalls); err != nil {
			return "", err
		}
		o.OllamaMsgs = append(o.OllamaMsgs, deepseekToolMessages(o.Tools.Results)...)
		o.Tools.reset()
		return o.SyncSend(ctx, l)
	}

	return response.Choices[0].Message.Content, nil
}

func GetDeepseekClient(ctx context.Context) *deepseek.Client {
	httpClient := utils.GetLLMProxyClient()
	txtType := utils.GetTxtType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	token := conf.BaseConfInfo.DeepseekToken
	if txtType == param.Ollama {
		token = "ollama"
	}
	client, err := deepseek.NewClientWithOptions(token, deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.ErrorCtx(ctx, "Error creating deepseek client", "err", err)
		return nil
	}

	if txtType == param.Ollama {
		client.Path = "api/chat"
		client.BaseURL = "http://localhost:11434/"
	}

	if conf.BaseConfInfo.CustomUrl != "" {
		client.BaseURL = conf.BaseConfInfo.CustomUrl
	}

	if conf.BaseConfInfo.CustomPath != "" {
		client.Path = conf.BaseConfInfo.CustomPath
	}

	return client
}

type Stream struct {
	resp   *http.Response
	reader *bufio.Reader
}

func requestDeepseek(ctx context.Context, c *deepseek.Client, request *deepseek.StreamChatCompletionRequest) (*Stream, error) {
	req, err := deepseekUtils.NewRequestBuilder(c.AuthToken).
		SetBaseURL(c.BaseURL).
		SetPath(c.Path).
		SetBodyFromStruct(request).
		BuildStream(ctx)

	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if resp == nil || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	return &Stream{
		resp:   resp,
		reader: bufio.NewReader(resp.Body),
	}, nil
}

func Receive(stream *Stream) (*deepseek.StreamChatCompletionResponse, error) {
	reader := stream.reader

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("error reading stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			trimmed := strings.TrimPrefix(line, "data: ")
			if trimmed == "[DONE]" {
				return nil, io.EOF
			}

			var resp deepseek.StreamChatCompletionResponse
			if err := json.Unmarshal([]byte(trimmed), &resp); err != nil {
				return nil, fmt.Errorf("unmarshal error (chatCompletion): %w, raw: %s", err, trimmed)
			}

			if resp.Usage == nil {
				resp.Usage = &deepseek.StreamUsage{}
			}

			return &resp, nil
		}

		var ollamaResp deepseek.OllamaStreamResponse
		if err := json.Unmarshal([]byte(line), &ollamaResp); err == nil && ollamaResp.Model != "" {
			resp := &deepseek.StreamChatCompletionResponse{
				Model: ollamaResp.Model,
				Choices: []deepseek.StreamChoices{
					{
						Index: 0,
						Delta: deepseek.StreamDelta{
							Content: ollamaResp.Message.Content,
							Role:    ollamaResp.Message.Role,
						},
						FinishReason: ollamaResp.DoneReason,
					},
				},
			}
			if ollamaResp.Done && ollamaResp.Message.Content == "" {
				return nil, io.EOF
			}
			return resp, nil
		}
	}
}

func (s *Stream) Close() {
	s.resp.Body.Close()
}

func init() {
	RegisterChatClient(param.Ollama, func() LLMClient { return &OllamaReq{} })
}
