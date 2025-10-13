package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
	"unicode"
	
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
)

type OllamaDeepseekReq struct {
	ToolCall           []deepseek.ToolCall
	ToolMessage        []deepseek.ChatCompletionMessage
	CurrentToolMessage []deepseek.ChatCompletionMessage
	
	DeepseekMsgs []deepseek.ChatCompletionMessage
}

func (d *OllamaDeepseekReq) GetModel(l *LLM) {
	l.Model = "deepseek-r1"
	userInfo, err := db.GetUserByID(l.UserId)
	if err != nil {
		logger.ErrorCtx(l.Ctx, "Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" {
		logger.InfoCtx(l.Ctx, "User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (d *OllamaDeepseekReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	metrics.AppRequestCount.WithLabelValues(l.Model).Inc()
	
	request := &deepseek.StreamChatCompletionRequest{
		Model:  l.Model,
		Stream: true,
		StreamOptions: deepseek.StreamOptions{
			IncludeUsage: true,
		},
		MaxTokens:        *conf.LLMConfInfo.MaxTokens,
		TopP:             float32(*conf.LLMConfInfo.TopP),
		FrequencyPenalty: float32(*conf.LLMConfInfo.FrequencyPenalty),
		TopLogProbs:      *conf.LLMConfInfo.TopLogProbs,
		LogProbs:         *conf.LLMConfInfo.LogProbs,
		Stop:             conf.LLMConfInfo.Stop,
		PresencePenalty:  float32(*conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(*conf.LLMConfInfo.Temperature),
	}
	
	request.Messages = d.DeepseekMsgs
	
	stream, err := deepseek.CreateOllamaChatCompletionStream(ctx, request)
	if err != nil {
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return err
	}
	defer stream.Close()
	msgInfoContent := &param.MsgInfo{
		SendLen: FirstSendLen,
	}
	
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	
	hasTools := false
	for {
		response, err := stream.Recv()
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
				err = d.requestToolsCall(ctx, choice, l)
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
			l.Token += response.Usage.TotalTokens
		}
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 {
		l.MessageChan <- msgInfoContent
	}
	
	if hasTools && len(d.CurrentToolMessage) != 0 {
		d.CurrentToolMessage = append([]deepseek.ChatCompletionMessage{
			{
				Role:      deepseek.ChatMessageRoleAssistant,
				Content:   l.WholeContent,
				ToolCalls: d.ToolCall,
			},
		}, d.CurrentToolMessage...)
		
		d.ToolMessage = append(d.ToolMessage, d.CurrentToolMessage...)
		d.DeepseekMsgs = append(d.DeepseekMsgs, d.CurrentToolMessage...)
		d.CurrentToolMessage = make([]deepseek.ChatCompletionMessage, 0)
		d.ToolCall = make([]deepseek.ToolCall, 0)
		return d.Send(ctx, l)
	}
	
	return nil
}

func (d *OllamaDeepseekReq) GetUserMessage(msg string) {
	d.GetMessage(constants.ChatMessageRoleUser, msg)
}

func (d *OllamaDeepseekReq) GetAssistantMessage(msg string) {
	d.GetMessage(constants.ChatMessageRoleAssistant, msg)
}

func (d *OllamaDeepseekReq) AppendMessages(client LLMClient) {
	if len(d.DeepseekMsgs) == 0 {
		d.DeepseekMsgs = make([]deepseek.ChatCompletionMessage, 0)
	}
	
	d.DeepseekMsgs = append(d.DeepseekMsgs, client.(*OllamaDeepseekReq).DeepseekMsgs...)
}

func (d *OllamaDeepseekReq) GetMessage(role, msg string) {
	if len(d.DeepseekMsgs) == 0 {
		d.DeepseekMsgs = []deepseek.ChatCompletionMessage{
			{
				Role:    role,
				Content: msg,
			},
		}
		return
	}
	
	d.DeepseekMsgs = append(d.DeepseekMsgs, deepseek.ChatCompletionMessage{
		Role:    role,
		Content: msg,
	})
}
func (d *OllamaDeepseekReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	start := time.Now()
	metrics.AppRequestCount.WithLabelValues(l.Model).Inc()
	
	request := &deepseek.ChatCompletionRequest{
		Model:            l.Model,
		MaxTokens:        *conf.LLMConfInfo.MaxTokens,
		TopP:             float32(*conf.LLMConfInfo.TopP),
		FrequencyPenalty: float32(*conf.LLMConfInfo.FrequencyPenalty),
		TopLogProbs:      *conf.LLMConfInfo.TopLogProbs,
		LogProbs:         *conf.LLMConfInfo.LogProbs,
		Stop:             conf.LLMConfInfo.Stop,
		PresencePenalty:  float32(*conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(*conf.LLMConfInfo.Temperature),
		Messages:         d.DeepseekMsgs,
		Tools:            l.DeepseekTools,
	}
	
	// assign task
	response, err := deepseek.CreateOllamaChatCompletion(request)
	
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	if err != nil {
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	if len(response.Choices) == 0 {
		logger.ErrorCtx(l.Ctx, "response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.DeepseekMsgs[len(d.DeepseekMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls, l)
		return d.SyncSend(ctx, l)
	}
	
	return response.Choices[0].Message.Content, nil
}

func (d *OllamaDeepseekReq) requestOneToolsCall(ctx context.Context, toolsCall []deepseek.ToolCall, l *LLM) {
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			return
		}
		
		toolsData, err := l.ExecMcpReq(ctx, tool.Function.Name, property)
		if err != nil {
			logger.WarnCtx(l.Ctx, "exec tools fail", "err", err)
			return
		}
		
		d.DeepseekMsgs = append(d.DeepseekMsgs, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
	}
}

func (d *OllamaDeepseekReq) requestToolsCall(ctx context.Context, choice deepseek.StreamChoices, l *LLM) error {
	
	for _, toolCall := range choice.Delta.ToolCalls {
		property := make(map[string]interface{})
		
		if toolCall.Function.Name != "" {
			d.ToolCall = append(d.ToolCall, toolCall)
			d.ToolCall[len(d.ToolCall)-1].Function.Name = toolCall.Function.Name
		}
		
		if toolCall.ID != "" {
			d.ToolCall[len(d.ToolCall)-1].ID = toolCall.ID
		}
		
		if toolCall.Type != "" {
			d.ToolCall[len(d.ToolCall)-1].Type = toolCall.Type
		}
		
		if toolCall.Function.Arguments != "" {
			d.ToolCall[len(d.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}
		
		err := json.Unmarshal([]byte(d.ToolCall[len(d.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}
		
		tool := d.ToolCall[len(d.ToolCall)-1]
		toolsData, err := l.ExecMcpReq(ctx, tool.Function.Name, property)
		if err != nil {
			logger.ErrorCtx(ctx, "Error executing MCP request", "toolId", toolCall.ID, "err", err)
			return err
		}
		d.CurrentToolMessage = append(d.CurrentToolMessage, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
	}
	
	return nil
	
}
