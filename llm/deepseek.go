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

type DeepseekReq struct {
	ToolCall           []deepseek.ToolCall
	ToolMessage        []deepseek.ChatCompletionMessage
	CurrentToolMessage []deepseek.ChatCompletionMessage
	
	DeepseekMsgs []deepseek.ChatCompletionMessage
}

// CallLLMAPI request DeepSeek API and get response
func (d *DeepseekReq) CallLLMAPI(ctx context.Context, prompt string, l *LLM) error {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	
	d.GetMessages(userId, prompt)
	
	logger.Info("msg receive", "userID", userId, "prompt", prompt)
	
	return d.Send(ctx, l)
}

func (d *DeepseekReq) GetModel(l *LLM) {
	_, _, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	
	l.Model = deepseek.DeepSeekChat
	userInfo, err := db.GetUserByID(userId)
	if err != nil {
		logger.Error("Error getting user info", "err", err)
	}
	if userInfo != nil && userInfo.Mode != "" && param.DeepseekModels[userInfo.Mode] {
		logger.Info("User info", "userID", userInfo.UserId, "mode", userInfo.Mode)
		l.Model = userInfo.Mode
	}
}

func (d *DeepseekReq) GetMessages(userId int64, prompt string) {
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

func (d *DeepseekReq) Send(ctx context.Context, l *LLM) error {
	if l.OverLoop() {
		return errors.New("too many loops")
	}
	
	start := time.Now()
	_, updateMsgID, userId := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	d.GetModel(l)
	
	// set deepseek proxy
	httpClient := utils.GetDeepseekProxyClient()
	
	client, err := deepseek.NewClientWithOptions(*conf.BaseConfInfo.DeepseekToken,
		deepseek.WithBaseURL(*conf.BaseConfInfo.CustomUrl), deepseek.WithHTTPClient(httpClient))
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
				err = d.requestToolsCall(ctx, choice)
				if err != nil {
					if errors.Is(err, ToolsJsonErr) {
						continue
					} else {
						logger.Error("requestToolsCall error", "updateMsgID", updateMsgID, "err", err)
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
		db.InsertMsgRecord(userId, &db.AQ{
			Question: l.Content,
			Answer:   l.WholeContent,
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
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(l.Update)
	
	d.GetModel(l)
	
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
		logger.Error("ChatCompletionStream error", "updateMsgID", updateMsgID, "err", err)
		return "", err
	}
	
	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return "", errors.New("response is empty")
	}
	
	l.Token += response.Usage.TotalTokens
	if len(response.Choices[0].Message.ToolCalls) > 0 {
		d.GetAssistantMessage("")
		d.DeepseekMsgs[len(d.DeepseekMsgs)-1].ToolCalls = response.Choices[0].Message.ToolCalls
		d.requestOneToolsCall(ctx, response.Choices[0].Message.ToolCalls)
	}
	
	return response.Choices[0].Message.Content, nil
}

func (d *DeepseekReq) requestOneToolsCall(ctx context.Context, toolsCall []deepseek.ToolCall) {
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			logger.Warn("json unmarshal fail", "err", err, "args", tool.Function.Arguments)
			return
		}
		
		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err, "name", tool.Function.Name, "args", property)
			return
		}
		
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err, "name", tool.Function.Name, "args", property)
			return
		}
		
		d.DeepseekMsgs = append(d.DeepseekMsgs, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
		logger.Info("exec tool", "name", tool.Function.Name, "toolsData", toolsData, "args", property)
	}
}

func (d *DeepseekReq) requestToolsCall(ctx context.Context, choice deepseek.StreamChoices) error {
	
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

// GetBalanceInfo get balance info
func GetBalanceInfo() *deepseek.BalanceResponse {
	client := deepseek.NewClient(*conf.BaseConfInfo.DeepseekToken)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	balance, err := deepseek.GetBalance(client, ctx)
	if err != nil {
		logger.Error("Error getting balance", "err", err)
	}
	
	if balance == nil || len(balance.BalanceInfos) == 0 {
		logger.Error("No balance information returned")
	}
	
	return balance
}
