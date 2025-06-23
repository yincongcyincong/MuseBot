package llm

import (
	"context"
	"encoding/json"
	"regexp"
	"time"
	
	"github.com/cohesion-org/deepseek-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
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
	
	llm := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan), WithContent(d.Content))
	
	prompt := i18n.GetMessage(*conf.Lang, "mcp_prompt", taskParam)
	llm.LLMClient.GetUserMessage(prompt)
	llm.Content = prompt
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("get message fail", "err", err)
		return
	}
	
	d.Token += llm.Token
	
	matches := jsonRe.FindAllString(c, -1)
	plans := new(TaskInfo)
	for _, match := range matches {
		err = json.Unmarshal([]byte(match), &plans)
		if err != nil {
			logger.Error("json umarshal fail", "err", err)
		}
	}
	
	if len(plans.Plan) == 0 {
		logger.Warn("no plan created!")
		
		finalLLM := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
			WithMessageChan(d.MessageChan), WithContent(d.Content))
		finalLLM.LLMClient.GetUserMessage(c)
		err = finalLLM.LLMClient.Send(ctx, finalLLM)
		if err != nil {
			logger.Warn("request summary fail", "err", err)
		}
		return
	}
	
	llm.LLMClient.GetAssistantMessage(c)
	summaryAQ := make(map[string]string)
	d.loopTask(ctx, plans, summaryAQ, c, llm)
	
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
	llm.LLMClient.GetUserMessage(i18n.GetMessage(*conf.Lang, "summary_task_prompt", summaryParam))
	err = llm.LLMClient.Send(ctx, llm)
	if err != nil {
		logger.Warn("request summary fail", "err", err)
	}
}

func (d *DeepseekTaskReq) loopTask(ctx context.Context, plans *TaskInfo,
	summaryAQ map[string]string, lastPlan string, llm *LLM) {
	summaryMsg := map[string]*LLM{}
	completeTasks := map[string]bool{}
	for _, plan := range plans.Plan {
		if _, ok := summaryMsg[plan.Name]; !ok {
			if _, ok := conf.TaskTools[plan.Name]; ok {
				summaryMsg[plan.Name] = NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
					WithMessageChan(d.MessageChan), WithContent(plan.Description), WithTaskTools(conf.TaskTools[plan.Name]))
			}
			
		}
		
		summaryMsg[plan.Name].LLMClient.GetUserMessage(plan.Description)
		
		logger.Info("execute task", "task", plan.Name)
		summaryAQ[plan.Description] = d.requestTask(ctx, summaryMsg, plan)
		completeTasks[plan.Description] = true
	}
	
	for _, su := range summaryMsg {
		d.Token += su.Token
	}
	
	taskParam := map[string]interface{}{
		"user_task":      d.Content,
		"complete_tasks": completeTasks,
		"last_plan":      lastPlan,
	}
	
	llm.LLMClient.GetUserMessage(i18n.GetMessage(*conf.Lang, "loop_task_prompt", taskParam))
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return
	}
	
	if len(c) == 0 {
		logger.Error("response is emtpy", "response", c)
		return
	}
	
	d.Token += llm.Token
	
	matches := jsonRe.FindAllString(c, -1)
	plans = new(TaskInfo)
	for _, match := range matches {
		err := json.Unmarshal([]byte(match), &plans)
		if err != nil {
			logger.Error("json umarshal fail", "err", err)
		}
	}
	
	llm.LLMClient.GetAssistantMessage(c)
	
	if len(plans.Plan) == 0 {
		return
	}
	
	d.loopTask(ctx, plans, summaryAQ, c, llm)
}

func (d *DeepseekTaskReq) requestTask(ctx context.Context, summaryMsg map[string]*LLM, plan *Task) string {
	
	c, err := summaryMsg[plan.Name].LLMClient.SyncSend(ctx, summaryMsg[plan.Name])
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return ""
	}
	
	// deepseek response merge into msg
	summaryMsg[plan.Name].LLMClient.GetAssistantMessage(c)
	
	return c
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
