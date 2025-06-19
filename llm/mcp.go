package llm

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

var (
	mcpRe = regexp.MustCompile(`\{\s*"agent"\s*:\s*"playwright"\s*\}`)
)

type McpResult struct {
	Agent string `json:"agent"`
}

func (d *DeepseekTaskReq) ExecuteMcp() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	d.Model = deepseek.DeepSeekChat
	_, updateMsgID, _ := utils.GetChatIdAndMsgIdAndUserID(d.Update)

	// set deepseek proxy
	httpClient := utils.GetDeepseekProxyClient()

	client, err := deepseek.NewClientWithOptions(*conf.DeepseekToken,
		deepseek.WithBaseURL(*conf.CustomUrl), deepseek.WithHTTPClient(httpClient))
	if err != nil {
		logger.Error("Error creating deepseek client", "err", err)
		return
	}

	logger.Info("mcp content", "content", d.Content)
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
			Content: i18n.GetMessage(*conf.Lang, "mcp_prompt", taskParam),
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

	matches := mcpRe.FindAllString(response.Choices[0].Message.Content, -1)
	mcpResult := new(McpResult)
	for _, match := range matches {
		err := json.Unmarshal([]byte(match), mcpResult)
		if err != nil {
			logger.Error("json umarshal fail", "err", err)
		}
	}

	tools := make([]deepseek.Tool, 0)
	if _, ok := conf.TaskTools[mcpResult.Agent]; ok {
		tools = conf.TaskTools[mcpResult.Agent].DeepseekTool
	}

	llm := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan), WithContent(d.Content),
		WithDeepseekTools(tools))

	err = llm.LLMClient.CallLLMAPI(ctx, d.Content, llm)
	if err != nil {
		logger.Error("execute conversation fail", "err", err)
	}
}
