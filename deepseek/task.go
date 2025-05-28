package deepseek

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/mcp-client-go/clients"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

var (
	jsonRe = regexp.MustCompile(`(\{\s*"plan":\s*\[\s*(?:\{\s*"name":\s*"[^"]*",\s*"description":\s*"[^"]*"\s*\}\s*,?\s*)+\]\s*\})`)
)

type DeepseekTaskReq struct {
	MessageChan chan *param.MsgInfo
	Update      tgbotapi.Update
	Bot         *tgbotapi.BotAPI
	Content     string
	Model       string
	Token       int
}

type Task struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TaskInfo struct {
	Plan []*Task `json:"plan"`
}

type TaskResult struct {
	TaskName   string
	TaskResult string
}

func (d *DeepseekTaskReq) ExecuteTask() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	d.Model = deepseek.DeepSeekChat
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(d.Update)

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
		return
	}

	logger.Info("task content", "content", d.Content)
	taskParam := make(map[string]interface{})
	taskParam["assign_param"] = make([]map[string]string, 0)
	taskParam["user_task"] = d.Content
	for name, tool := range conf.TaskTools {
		taskParam["assign_param"] = append(taskParam["assign_param"].([]map[string]string), map[string]string{
			"tool_name": name,
			"tool_desc": tool.Description,
		})
	}

	messages := []deepseek.ChatCompletionMessage{
		{
			Role:    constants.ChatMessageRoleUser,
			Content: i18n.GetMessage(*conf.Lang, "assign_task_prompt", taskParam),
		},
	}

	request := &deepseek.ChatCompletionRequest{
		Model:            d.Model,
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
		Messages:         messages,
	}

	// assign task
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "updateMsgID", updateMsgID, "err", err)
		return
	}

	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return
	}

	matches := jsonRe.FindAllString(response.Choices[0].Message.Content, -1)
	plans := new(TaskInfo)
	for _, match := range matches {
		err = json.Unmarshal([]byte(match), &plans)
		if err != nil {
			logger.Error("json umarshal fail", "err", err)
		}
	}

	if len(plans.Plan) == 0 {
		logger.Error("no plan created!")
		return
	}

	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleAssistant,
		Content: response.Choices[0].Message.Content,
	})
	summaryAQ := make(map[string]string)
	messages = d.loopTask(ctx, plans, client, messages, summaryAQ, response.Choices[0].Message.Content)

	// summary
	summaryParam := make(map[string]interface{})
	summaryParam["aq"] = make([]map[string]string, 0)
	summaryParam["user_task"] = d.Content
	for t, a := range summaryAQ {
		summaryParam["aq"] = append(summaryParam["aq"].([]map[string]string), map[string]string{
			"task":   t,
			"answer": a,
		})
	}
	summaryDPMsg := []deepseek.ChatCompletionMessage{
		{
			Role:    constants.ChatMessageRoleUser,
			Content: i18n.GetMessage(*conf.Lang, "summary_task_prompt", summaryParam),
		},
	}
	err = d.send(ctx, append(messages, summaryDPMsg...))
	if err != nil {
		logger.Warn("request summary fail", "err", err)
	}
}

func (d *DeepseekTaskReq) loopTask(ctx context.Context, plans *TaskInfo, client *deepseek.Client,
	messages []deepseek.ChatCompletionMessage, summaryAQ map[string]string, lastPlan string) []deepseek.ChatCompletionMessage {
	summaryMsg := map[string][]deepseek.ChatCompletionMessage{}
	completeTasks := map[string]bool{}
	for _, plan := range plans.Plan {
		if _, ok := summaryMsg[plan.Name]; !ok {
			summaryMsg[plan.Name] = make([]deepseek.ChatCompletionMessage, 0)
		}

		summaryMsg[plan.Name] = append(summaryMsg[plan.Name], deepseek.ChatCompletionMessage{
			Role:    constants.ChatMessageRoleUser,
			Content: plan.Description,
		})

		logger.Info("execute task", "task", plan.Name)
		summaryAQ[plan.Description] = d.requestTask(ctx, summaryMsg, client, plan)
		completeTasks[plan.Description] = true
	}

	taskParam := map[string]interface{}{
		"user_task":      d.Content,
		"complete_tasks": completeTasks,
		"last_plan":      lastPlan,
	}

	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: i18n.GetMessage(*conf.Lang, "loop_task_prompt", taskParam),
	})

	request := &deepseek.ChatCompletionRequest{
		Model:            d.Model,
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
		Messages:         messages,
	}

	// assign task
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return messages
	}

	if len(response.Choices) == 0 {
		logger.Error("response is emtpy", "response", response)
		return messages
	}

	matches := jsonRe.FindAllString(response.Choices[0].Message.Content, -1)
	plans = new(TaskInfo)
	for _, match := range matches {
		err := json.Unmarshal([]byte(match), &plans)
		if err != nil {
			logger.Error("json umarshal fail", "err", err)
		}
	}

	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleAssistant,
		Content: response.Choices[0].Message.Content,
	})

	if len(plans.Plan) == 0 {
		return messages
	}

	return d.loopTask(ctx, plans, client, messages, summaryAQ, response.Choices[0].Message.Content)
}

func (d *DeepseekTaskReq) requestTask(ctx context.Context, summaryMsg map[string][]deepseek.ChatCompletionMessage,
	client *deepseek.Client, plan *Task) string {
	var tools []deepseek.Tool
	if _, ok := conf.TaskTools[plan.Name]; ok {
		tools = conf.TaskTools[plan.Name].DeepseekTool
	}

	request := &deepseek.ChatCompletionRequest{
		Model:            d.Model,
		MaxTokens:        *conf.MaxTokens,
		TopP:             float32(*conf.TopP),
		FrequencyPenalty: float32(*conf.FrequencyPenalty),
		TopLogProbs:      *conf.TopLogProbs,
		LogProbs:         *conf.LogProbs,
		Stop:             conf.Stop,
		PresencePenalty:  float32(*conf.PresencePenalty),
		Temperature:      float32(*conf.Temperature),
		Tools:            tools,
		Messages:         summaryMsg[plan.Name],
	}

	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return ""
	}

	hasTools := false
	msgContent := ""
	for _, choice := range response.Choices {
		if len(choice.Message.ToolCalls) > 0 {
			summaryMsg[plan.Name] = d.requestToolsCall(ctx, choice.Message.ToolCalls, summaryMsg[plan.Name])
			hasTools = true
		}
		msgContent += choice.Message.Content
	}

	if hasTools {
		return d.requestTask(ctx, summaryMsg, client, plan)
	}

	// deepseek response merge into msg
	summaryMsg[plan.Name] = append(summaryMsg[plan.Name], deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleAssistant,
		Content: msgContent,
	})

	return msgContent
}

func (d *DeepseekTaskReq) requestToolsCall(ctx context.Context, toolsCall []deepseek.ToolCall,
	msg []deepseek.ChatCompletionMessage) []deepseek.ChatCompletionMessage {
	msg = append(msg, deepseek.ChatCompletionMessage{
		Role:      deepseek.ChatMessageRoleAssistant,
		Content:   "",
		ToolCalls: toolsCall,
	})
	for _, tool := range toolsCall {
		property := make(map[string]interface{})
		err := json.Unmarshal([]byte(tool.Function.Arguments), &property)
		if err != nil {
			return nil
		}

		mc, err := clients.GetMCPClientByToolName(tool.Function.Name)
		if err != nil {
			logger.Warn("get mcp fail", "err", err)
			return nil
		}

		logger.Info("exec tool", "name", tool.Function.Name)
		toolsData, err := mc.ExecTools(ctx, tool.Function.Name, property)
		if err != nil {
			logger.Warn("exec tools fail", "err", err)
			return nil
		}
		msg = append(msg, deepseek.ChatCompletionMessage{
			Role:       constants.ChatMessageRoleTool,
			Content:    toolsData,
			ToolCallID: tool.ID,
		})
	}

	return msg
}

func (d *DeepseekTaskReq) send(ctx context.Context, messages []deepseek.ChatCompletionMessage) error {
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(d.Update)
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
		Model:  d.Model,
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
			d.sendMsg(msgInfoContent, choice)
		}

		if response.Usage != nil {
			d.Token += response.Usage.TotalTokens
			metrics.TotalTokens.Add(float64(d.Token))
		}
	}

	d.MessageChan <- msgInfoContent

	return nil
}

func (d *DeepseekTaskReq) sendMsg(msgInfoContent *param.MsgInfo, choice deepseek.StreamChoices) {
	// exceed max telegram one message length
	if utils.Utf16len(msgInfoContent.Content) > OneMsgLen {
		d.MessageChan <- msgInfoContent
		msgInfoContent = &param.MsgInfo{
			SendLen: NonFirstSendLen,
		}
	}

	msgInfoContent.Content += choice.Delta.Content
	if len(msgInfoContent.Content) > msgInfoContent.SendLen {
		d.MessageChan <- msgInfoContent
		msgInfoContent.SendLen += NonFirstSendLen
	}
}
