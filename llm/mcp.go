package llm

import (
	"context"
	"encoding/json"
	"regexp"
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

var (
	mcpRe = regexp.MustCompile(`\{\s*"agent"\s*:\s*"[^\"]*"\s*\}`)
)

type McpResult struct {
	Agent string `json:"agent"`
}

// ExecuteMcp execute mcp request
func (d *DeepseekTaskReq) ExecuteMcp() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	
	logger.Info("mcp content", "content", d.Content)
	taskParam := make(map[string]interface{})
	taskParam["assign_param"] = make([]map[string]string, 0)
	taskParam["user_task"] = d.Content
	conf.TaskTools.Range(func(name, value any) bool {
		tool := value.(*conf.AgentInfo)
		taskParam["assign_param"] = append(taskParam["assign_param"].([]map[string]string), map[string]string{
			"tool_name": name.(string),
			"tool_desc": tool.Description,
		})
		return true
	})
	
	// get mcp request
	chatId, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(d.Update)
	llm := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan), WithContent(d.Content))
	
	prompt := i18n.GetMessage(*conf.BaseConfInfo.Lang, "mcp_prompt", taskParam)
	llm.LLMClient.GetUserMessage(prompt)
	llm.Content = prompt
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("get message fail", "err", err)
		utils.SendMsg(chatId, err.Error(), d.Bot, msgId, "")
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
	var taskTool *conf.AgentInfo
	taskToolInter, ok := conf.TaskTools.Load(mcpResult.Agent)
	if ok {
		taskTool = taskToolInter.(*conf.AgentInfo)
	}
	mcpLLM := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan), WithContent(d.Content), WithTaskTools(taskTool))
	mcpLLM.Token += llm.Token
	mcpLLM.Content = d.Content
	mcpLLM.LLMClient.GetUserMessage(d.Content)
	err = mcpLLM.LLMClient.Send(ctx, mcpLLM)
	if err != nil {
		utils.SendMsg(chatId, err.Error(), d.Bot, msgId, "")
		logger.Error("execute conversation fail", "err", err)
	}
}
