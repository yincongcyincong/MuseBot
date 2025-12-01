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
	"unicode"
	
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
	ToolCall           []deepseek.ToolCall
	ToolMessage        []deepseek.ChatCompletionMessage
	CurrentToolMessage []deepseek.ChatCompletionMessage
	
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

func (o OllamaReq) Send(ctx context.Context, l *LLM) error {
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
	
	if conf.BaseConfInfo.LLMOptionParam {
		request.MaxTokens = conf.LLMConfInfo.MaxTokens
		request.TopP = float32(conf.LLMConfInfo.TopP)
		request.FrequencyPenalty = float32(conf.LLMConfInfo.FrequencyPenalty)
		request.TopLogProbs = conf.LLMConfInfo.TopLogProbs
		request.LogProbs = conf.LLMConfInfo.LogProbs
		request.Stop = conf.LLMConfInfo.Stop
		request.PresencePenalty = float32(conf.LLMConfInfo.PresencePenalty)
		request.Temperature = float32(conf.LLMConfInfo.Temperature)
	}
	
	request.Messages = o.OllamaMsgs
	
	var stream *Stream
	var err error
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		stream, err = requestDeepseek(ctx, client, request)
		if err != nil {
			logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
			continue
		}
		break
	}
	if err != nil || stream == nil {
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return fmt.Errorf("request fail %v %v", err, stream)
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}
	
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	
	hasTools := false
	for {
		response, err := Receive(stream)
		if errors.Is(err, io.EOF) {
			logger.InfoCtx(l.Ctx, "Stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.WarnCtx(l.Ctx, "Stream error", "updateMsgID", l.MsgId, "err", err)
			return err
		}
		for _, choice := range response.Choices {
			if len(choice.Delta.ToolCalls) > 0 {
				hasTools = true
				err = o.RequestToolsCall(ctx, choice, l)
				if err != nil {
					if errors.Is(err, ToolsJsonErr) {
						continue
					} else {
						logger.ErrorCtx(l.Ctx, "requestToolsCall error", "updateMsgID", l.MsgId, "err", err)
					}
				}
			}
			
			if !hasTools {
				msgInfoContent = l.SendMsg(msgInfoContent, choice.Delta.Content)
			}
		}
		
		if response.Usage != nil {
			l.Cs.Token += response.Usage.TotalTokens
		}
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 || (hasTools && conf.BaseConfInfo.SendMcpRes) {
		if conf.BaseConfInfo.Powered != "" {
			msgInfoContent.Content = msgInfoContent.Content + "\n\n" + conf.BaseConfInfo.Powered
		}
		l.MessageChan <- msgInfoContent
	}
	
	if hasTools && len(o.CurrentToolMessage) != 0 {
		o.CurrentToolMessage = append([]deepseek.ChatCompletionMessage{
			{
				Role:      deepseek.ChatMessageRoleAssistant,
				Content:   l.WholeContent,
				ToolCalls: o.ToolCall,
			},
		}, o.CurrentToolMessage...)
		
		o.ToolMessage = append(o.ToolMessage, o.CurrentToolMessage...)
		o.OllamaMsgs = append(o.OllamaMsgs, o.CurrentToolMessage...)
		o.CurrentToolMessage = make([]deepseek.ChatCompletionMessage, 0)
		o.ToolCall = make([]deepseek.ToolCall, 0)
		return o.Send(ctx, l)
	}
	
	return nil
}

func (o OllamaReq) GetUserMessage(msg string) {
	o.GetMessage(constants.ChatMessageRoleUser, msg)
}

func (o OllamaReq) GetAssistantMessage(msg string) {
	o.GetMessage(constants.ChatMessageRoleAssistant, msg)
}

func (o OllamaReq) GetSystemMessage(msg string) {
	o.GetMessage(constants.ChatMessageRoleSystem, msg)
}

func (o OllamaReq) GetImageMessage(image [][]byte, msg string) {}

func (o OllamaReq) GetAudioMessage(audio []byte, msg string) {}

func (o OllamaReq) AppendMessages(client LLMClient) {
	if len(o.OllamaMsgs) == 0 {
		o.OllamaMsgs = make([]deepseek.ChatCompletionMessage, 0)
	}
	
	o.OllamaMsgs = append(o.OllamaMsgs, client.(*OllamaReq).OllamaMsgs...)
}

func (o OllamaReq) GetMessage(role, msg string) {
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

func (o OllamaReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	
	start := time.Now()
	
	client := GetDeepseekClient(ctx)
	request := &deepseek.ChatCompletionRequest{
		Model:    l.Model,
		Messages: o.OllamaMsgs,
		Tools:    l.DeepseekTools,
	}
	
	if conf.BaseConfInfo.LLMOptionParam {
		request.MaxTokens = conf.LLMConfInfo.MaxTokens
		request.TopP = float32(conf.LLMConfInfo.TopP)
		request.FrequencyPenalty = float32(conf.LLMConfInfo.FrequencyPenalty)
		request.TopLogProbs = conf.LLMConfInfo.TopLogProbs
		request.LogProbs = conf.LLMConfInfo.LogProbs
		request.Stop = conf.LLMConfInfo.Stop
		request.PresencePenalty = float32(conf.LLMConfInfo.PresencePenalty)
		request.Temperature = float32(conf.LLMConfInfo.Temperature)
	}
	
	// assign task
	var response *deepseek.ChatCompletionResponse
	var err error
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		response, err = client.CreateChatCompletion(ctx, request)
		if err != nil {
			logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
			continue
		}
		break
	}
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
		o.GetAssistantMessage("")
		o.OllamaMsgs[len(o.OllamaMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		o.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls, l)
		return o.SyncSend(ctx, l)
	}
	
	return response.Choices[0].Message.Content, nil
}

func (o OllamaReq) requestOneToolsCall(ctx context.Context, toolsCall []deepseek.ToolCall, l *LLM) {
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			logger.WarnCtx(l.Ctx, "json unmarshal fail", "err", err, "args", tool.Function.Arguments)
			return
		}
		
		toolsData, err := l.ExecMcpReq(ctx, tool.Function.Name, property)
		if err != nil {
			logger.WarnCtx(l.Ctx, "exec tools fail", "err", err, "name", tool.Function.Name, "args", property)
			return
		}
		
		o.OllamaMsgs = append(o.OllamaMsgs, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
	}
}

func (o OllamaReq) RequestToolsCall(ctx context.Context, choice deepseek.StreamChoices, l *LLM) error {
	
	for _, toolCall := range choice.Delta.ToolCalls {
		property := make(map[string]interface{})
		
		if toolCall.Function.Name != "" {
			o.ToolCall = append(o.ToolCall, toolCall)
			o.ToolCall[len(o.ToolCall)-1].Function.Name = toolCall.Function.Name
		}
		
		if toolCall.ID != "" {
			o.ToolCall[len(o.ToolCall)-1].ID = toolCall.ID
		}
		
		if toolCall.Type != "" {
			o.ToolCall[len(o.ToolCall)-1].Type = toolCall.Type
		}
		
		if toolCall.Function.Arguments != "" && toolCall.Function.Name == "" {
			o.ToolCall[len(o.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}
		
		err := json.Unmarshal([]byte(o.ToolCall[len(o.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}
		
		tool := o.ToolCall[len(o.ToolCall)-1]
		toolsData, err := l.ExecMcpReq(ctx, tool.Function.Name, property)
		if err != nil {
			logger.ErrorCtx(ctx, "Error executing MCP request", "toolId", tool.ID, "err", err)
			return err
		}
		
		o.CurrentToolMessage = append(o.CurrentToolMessage, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
	}
	
	return nil
	
}

func GetDeepseekClient(ctx context.Context) *deepseek.Client {
	httpClient := utils.GetLLMProxyClient()
	txtType := utils.GetTxtType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	if txtType == param.Ollama {
		conf.BaseConfInfo.DeepseekToken = "ollama"
	}
	client, err := deepseek.NewClientWithOptions(conf.BaseConfInfo.DeepseekToken, deepseek.WithHTTPClient(httpClient))
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
