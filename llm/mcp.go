package llm

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
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

	// get mcp request
	llm := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan), WithContent(d.Content))

	prompt := i18n.GetMessage(*conf.Lang, "mcp_prompt", taskParam)
	llm.LLMClient.GetMessage(prompt)
	llm.Content = prompt
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("get message fail", "err", err)
		return
	}

	matches := mcpRe.FindAllString(c, -1)
	mcpResult := new(McpResult)
	for _, match := range matches {
		err := json.Unmarshal([]byte(match), mcpResult)
		if err != nil {
			logger.Error("json umarshal fail", "err", err)
		}
	}

	// execute mcp request
	taskTool := conf.TaskTools[mcpResult.Agent]
	llm = NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan), WithContent(d.Content), WithTaskTools(taskTool))
	err = llm.LLMClient.CallLLMAPI(ctx, d.Content, llm)
	if err != nil {
		logger.Error("execute conversation fail", "err", err)
	}
}
