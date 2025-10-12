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
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"github.com/yincongcyincong/mcp-client-go/clients"
)

type DeepseekReq struct {
	ToolCall           []deepseek.ToolCall
	ToolMessage        []deepseek.ChatCompletionMessage
	CurrentToolMessage []deepseek.ChatCompletionMessage
	
	DeepseekMsgs []deepseek.ChatCompletionMessage
}

func (d *DeepseekReq) GetModel(l *LLM) {
	l.Model = deepseek.DeepSeekChat
	userInfo, err := db.GetUserByID(l.UserId)
	if err != nil {
		logger.ErrorCtx(l.Ctx, "Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.DeepseekModels[userInfo.Mode] {
		logger.InfoCtx(l.Ctx, "User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (d *DeepseekReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	start := time.Now()
	
	// set deepseek proxy
	metrics.APIRequestCount.WithLabelValues(l.Model).Inc()
	httpClient := utils.GetLLMProxyClient()
	client, err := deepseek.NewClientWithOptions(*conf.BaseConfInfo.DeepseekToken, deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.ErrorCtx(l.Ctx, "Error creating deepseek client", "err", err)
		return err
	}
	
	if *conf.BaseConfInfo.CustomUrl != "" {
		client.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	
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
		Tools:            l.DeepseekTools,
	}
	
	request.Messages = d.DeepseekMsgs
	
	stream, err := client.CreateChatCompletionStream(ctx, request)
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
				err = d.RequestToolsCall(ctx, choice, l)
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

func (d *DeepseekReq) GetUserMessage(msg string) {
	d.GetMessage(constants.ChatMessageRoleUser, msg)
}

func (d *DeepseekReq) GetAssistantMessage(msg string) {
	d.GetMessage(constants.ChatMessageRoleAssistant, msg)
}

func (d *DeepseekReq) AppendMessages(client LLMClient) {
	if len(d.DeepseekMsgs) == 0 {
		d.DeepseekMsgs = make([]deepseek.ChatCompletionMessage, 0)
	}
	
	d.DeepseekMsgs = append(d.DeepseekMsgs, client.(*DeepseekReq).DeepseekMsgs...)
}

func (d *DeepseekReq) GetMessage(role, msg string) {
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

func (d *DeepseekReq) SyncSend(ctx context.Context, l *LLM) (string, error) {
	metrics.APIRequestCount.WithLabelValues(l.Model).Inc()
	start := time.Now()
	
	httpClient := utils.GetLLMProxyClient()
	client, err := deepseek.NewClientWithOptions(*conf.BaseConfInfo.DeepseekToken, deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.ErrorCtx(l.Ctx, "Error creating deepseek client", "err", err)
		return "", err
	}
	
	if *conf.BaseConfInfo.CustomUrl != "" {
		client.BaseURL = *conf.BaseConfInfo.CustomUrl
	}
	
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
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.ErrorCtx(l.Ctx, "ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
	
	if len(response.Choices) == 0 {
		logger.ErrorCtx(l.Ctx, "response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	l.Token += response.Usage.TotalTokens
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.DeepseekMsgs[len(d.DeepseekMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls, l)
		return d.SyncSend(ctx, l)
	}
	
	return response.Choices[0].Message.Content, nil
}

func (d *DeepseekReq) requestOneToolsCall(ctx context.Context, toolsCall []deepseek.ToolCall, l *LLM) {
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			logger.WarnCtx(l.Ctx, "json unmarshal fail", "err", err, "args", tool.Function.Arguments)
			return
		}
		
		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.WarnCtx(l.Ctx, "get mcp fail", "err", err, "name", tool.Function.Name, "args", property)
			return
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.WarnCtx(l.Ctx, "exec tools fail", "err", err, "name", tool.Function.Name, "args", property)
			return
		}
		
		d.DeepseekMsgs = append(d.DeepseekMsgs, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
		logger.InfoCtx(l.Ctx, "exec tool", "name", tool.Function.Name, "toolsData", toolsData, "args", property)
		l.DirectSendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "send_mcp_info", map[string]interface{}{
			"function_name": tool.Function.Name,
			"request_args":  property,
			"response":      toolsData,
		}))
	}
}

func (d *DeepseekReq) RequestToolsCall(ctx context.Context, choice deepseek.StreamChoices, l *LLM) error {
	
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
		
		if toolCall.Function.Arguments != "" && toolCall.Function.Name == "" {
			d.ToolCall[len(d.ToolCall)-1].Function.Arguments += toolCall.Function.Arguments
		}
		
		err := json.Unmarshal([]byte(d.ToolCall[len(d.ToolCall)-1].Function.Arguments), &property)
		if err != nil {
			return ToolsJsonErr
		}
		
		tool := d.ToolCall[len(d.ToolCall)-1]
		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.WarnCtx(l.Ctx, "get mcp fail", "err", err, "function", tool.Function.Name,
				"toolCall", tool.ID, "argument", tool.Function.Arguments)
			return err
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.WarnCtx(l.Ctx, "exec tools fail", "err", err, "function", tool.Function.Name,
				"toolCall", tool.ID, "argument", tool.Function.Arguments)
			return err
		}
		d.CurrentToolMessage = append(d.CurrentToolMessage, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
		logger.InfoCtx(l.Ctx, "send tool request", "function", tool.Function.Name,
			"toolCall", tool.ID, "argument", tool.Function.Arguments,
			"res", toolsData)
		l.DirectSendMsg(i18n.GetMessage(*conf.BaseConfInfo.Lang, "send_mcp_info", map[string]interface{}{
			"function_name": tool.Function.Name,
			"request_args":  property,
			"response":      toolsData,
		}))
	}
	
	return nil
	
}

// GetBalanceInfo get balance info
func GetBalanceInfo(ctx context.Context) *deepseek.BalanceResponse {
	client := deepseek.NewClient(*conf.BaseConfInfo.DeepseekToken)
	balance, err := deepseek.GetBalance(client, ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Error getting balance", "err", err)
	}
	
	if balance == nil || len(balance.BalanceInfos) == 0 {
		logger.ErrorCtx(ctx, "No balance information returned")
	}
	
	return balance
}
