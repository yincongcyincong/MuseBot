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
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/db"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type OllamaDeepseekReq struct {
	ToolCall           []deepseek.ToolCall
	ToolMessage        []deepseek.ChatCompletionMessage
	CurrentToolMessage []deepseek.ChatCompletionMessage
	
	DeepseekMsgs []deepseek.ChatCompletionMessage
}

// CallLLMAPI request DeepSeek API and get response
func (d *OllamaDeepseekReq) CallLLMAPI(ctx context.Context, l *LLM) error {
	d.GetMessages(l.UserId, l.Content)
	
	logger.Info("msg receive", "userID", l.UserId, "prompt", l.Content)
	
	return d.Send(ctx, l)
}

func (d *OllamaDeepseekReq) GetModel(l *LLM) {
	l.Model = "llava:latest"
}

func (d *OllamaDeepseekReq) GetMessages(userId int64, prompt string) {
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
					err := json.Unmarshal([]byte(record.Content), &toolsMsgs)
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
	
	d.DeepseekMsgs = messages
}

func (d *OllamaDeepseekReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	
	request := &deepseek.StreamChatCompletionRequest{
		Model:  "llava:latest",
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
		logger.Error("ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
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
			logger.Info("Stream finished", "updateMsgID", l.MsgId)
			break
		}
		if err != nil {
			logger.Warn("Stream error", "updateMsgID", l.MsgId, "err", err)
			break
		}
		for _, choice := range response.Choices {
			if len(choice.Delta.ToolCalls) > 0 {
				hasTools = true
				err = d.requestToolsCall(ctx, choice)
				if err != nil {
					if errors.Is(err, ToolsJsonErr) {
						continue
					} else {
						logger.Error("requestToolsCall error", "updateMsgID", l.MsgId, "err", err)
					}
				}
			}
			
			if len(choice.Delta.Content) > 0 {
				msgInfoContent = l.sendMsg(msgInfoContent, choice.Delta.Content)
			}
		}
		
		if response.Usage != nil {
			l.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(l.Token))
		}
	}
	
	if l.MessageChan != nil && len(strings.TrimRightFunc(msgInfoContent.Content, unicode.IsSpace)) > 0 {
		l.MessageChan <- msgInfoContent
	}
	
	if !hasTools || len(d.CurrentToolMessage) == 0 {
		data, _ := json.Marshal(d.ToolMessage)
		db.InsertMsgRecord(l.UserId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
			Content:  string(data),
			Token:    l.Token,
		}, true)
	} else {
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
	
	// record time costing in dialog
	totalDuration := time.Since(start).Seconds()
	metrics.ConversationDuration.Observe(totalDuration)
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
	
	httpClient := utils.GetDeepseekProxyClient()
	
	client, err := deepseek.NewClientWithOptions(*conf.BaseConfInfo.DeepseekToken,
		deepseek.WithBaseURL(*conf.BaseConfInfo.CustomUrl), deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.Error("Error creating deepseek client", "err", err)
		return "", err
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
		logger.Error("ChatCompletionStream error", "updateMsgID", l.MsgId, "err", err)
		return "", err
	}
	
	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.DeepseekMsgs[len(d.DeepseekMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls)
	}
	
	return response.Choices[0].Message.Content, nil
}

func (d *OllamaDeepseekReq) requestOneToolsCall(ctx context.Context, toolsCall []deepseek.ToolCall) {
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			return
		}
		
		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return
		}
		
		d.DeepseekMsgs = append(d.DeepseekMsgs, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
		logger.Info("exec tool", "name", tool.Function.Name, "toolsData", toolsData)
	}
}

func (d *OllamaDeepseekReq) requestToolsCall(ctx context.Context, choice deepseek.StreamChoices) error {
	
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
		
		mc, err := clients.GetMCPClientByToolName(d.ToolCall[len(d.ToolCall)-1].Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err, "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
				"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments)
			return err
		}
		
		toolsData, err := mc.ExecTools(ctx, d.ToolCall[len(d.ToolCall)-1].Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err, "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
				"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments)
			return err
		}
		d.CurrentToolMessage = append(d.CurrentToolMessage, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: d.ToolCall[len(d.ToolCall)-1].ID,
		})
		
		logger.Info("send tool request", "function", d.ToolCall[len(d.ToolCall)-1].Function.Name,
			"toolCall", d.ToolCall[len(d.ToolCall)-1].ID, "argument", d.ToolCall[len(d.ToolCall)-1].Function.Arguments,
			"res", toolsData)
	}
	
	return nil
	
}
