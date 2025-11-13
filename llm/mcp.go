package llm

import (
	"encoding/json"
	"regexp"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
)

var (
	mcpRe = regexp.MustCompile(`\{\s*"agent"\s*:\s*"[^\"]*"\s*\}`)
)

type McpResult struct {
	Agent string `json:"agent"`
}

// ExecuteMcp execute mcp request
func (d *LLMTaskReq) ExecuteMcp() error {
	logger.InfoCtx(d.Ctx, "mcp content", "content", d.Content)
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
	llm := NewLLM(WithChatId(d.ChatId), WithMsgId(d.MsgId), WithUserId(d.UserId),
		WithMessageChan(d.MessageChan), WithContent(d.Content), WithHTTPMsgChan(d.HTTPMsgChan),
		WithPerMsgLen(d.PerMsgLen), WithContext(d.Ctx))
	
	prompt := i18n.GetMessage("mcp_prompt", taskParam)
	llm.LLMClient.GetUserMessage(prompt)
	llm.Content = prompt
	llm.LLMClient.GetModel(llm)
	
	metrics.APIRequestCount.WithLabelValues(llm.Model).Inc()
	c, err := llm.LLMClient.SyncSend(d.Ctx, llm)
	if err != nil {
		logger.ErrorCtx(d.Ctx, "get message fail", "err", err)
		return err
	}
	
	matches := mcpRe.FindAllString(c, -1)
	mcpResult := new(McpResult)
	for _, match := range matches {
		err := json.Unmarshal([]byte(match), mcpResult)
		if err != nil {
			logger.ErrorCtx(d.Ctx, "json umarshal fail", "err", err)
		}
	}
	
	llm.DirectSendMsg(c)
	logger.InfoCtx(d.Ctx, "mcp plan", "plan", mcpResult)
	
	// execute mcp request
	var taskTool *conf.AgentInfo
	taskToolInter, ok := conf.TaskTools.Load(mcpResult.Agent)
	if ok {
		taskTool = taskToolInter.(*conf.AgentInfo)
	}
	mcpLLM := NewLLM(WithChatId(d.ChatId), WithMsgId(d.MsgId), WithUserId(d.UserId),
		WithMessageChan(d.MessageChan), WithContent(d.Content), WithTaskTools(taskTool),
		WithPerMsgLen(d.PerMsgLen), WithContext(d.Ctx))
	mcpLLM.Cs.Token += llm.Cs.Token
	mcpLLM.Content = d.Content
	mcpLLM.LLMClient.GetUserMessage(d.Content)
	mcpLLM.LLMClient.GetModel(mcpLLM)
	
	metrics.APIRequestCount.WithLabelValues(mcpLLM.Model).Inc()
	err = mcpLLM.LLMClient.Send(d.Ctx, mcpLLM)
	if err != nil {
		logger.ErrorCtx(d.Ctx, "execute conversation fail", "err", err)
		return err
	}
	
	err = mcpLLM.InsertOrUpdate()
	if err != nil {
		logger.ErrorCtx(d.Ctx, "insertOrUpdate fail", "err", err)
	}
	
	return err
}
