package llm

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/i18n"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

var (
	jsonRe = regexp.MustCompile(`(?s)\{[\s\r\n]*"plan"\s*:\s*\[.*?][\s\r\n]*}`)
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

// ExecuteTask execute task command
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
	
	chatId, msgId, _ := utils.GetChatIdAndMsgIdAndUserID(d.Update)
	prompt := i18n.GetMessage(*conf.BaseConfInfo.Lang, "assign_task_prompt", taskParam)
	llm := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan))
	llm.LLMClient.GetUserMessage(prompt)
	llm.Content = prompt
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("get message fail", "err", err)
		utils.SendMsg(chatId, err.Error(), d.Bot, msgId, "")
		return
	}
	
	d.Token += llm.Token
	
	matches := jsonRe.FindAllString(c, -1)
	plans := new(TaskInfo)
	for _, match := range matches {
		err = json.Unmarshal([]byte(match), &plans)
		if err != nil {
			logger.Warn("json umarshal fail", "err", err)
		}
	}
	
	if len(plans.Plan) == 0 {
		logger.Info("no plan created!")
		
		finalLLM := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
			WithMessageChan(d.MessageChan), WithContent(d.Content))
		finalLLM.LLMClient.GetUserMessage(c)
		err = finalLLM.LLMClient.Send(ctx, finalLLM)
		if err != nil {
			logger.Error("request summary fail", "err", err)
			utils.SendMsg(chatId, err.Error(), d.Bot, msgId, "")
		}
		return
	}
	
	llm.LLMClient.GetAssistantMessage(c)
	err = d.loopTask(ctx, plans, c, llm, 0)
	if err != nil {
		logger.Error("loopTask fail", "err", err)
		utils.SendMsg(chatId, err.Error(), d.Bot, msgId, "")
		return
	}
	
	// summary
	summaryParam := make(map[string]interface{})
	summaryParam["user_task"] = d.Content
	summaryPrompt := i18n.GetMessage(*conf.BaseConfInfo.Lang, "summary_task_prompt", summaryParam)
	llm.LLMClient.GetUserMessage(summaryPrompt)
	llm.Content = summaryPrompt
	err = llm.LLMClient.Send(ctx, llm)
	if err != nil {
		logger.Error("request summary fail", "err", err)
		utils.SendMsg(chatId, err.Error(), d.Bot, msgId, "")
	}
}

// loopTask loop task
func (d *DeepseekTaskReq) loopTask(ctx context.Context, plans *TaskInfo, lastPlan string, llm *LLM, loop int) error {
	if loop > MostLoop {
		return errors.New("too many loops")
	}
	
	completeTasks := map[string]bool{}
	taskLLM := NewLLM(WithBot(d.Bot), WithUpdate(d.Update),
		WithMessageChan(d.MessageChan))
	for _, plan := range plans.Plan {
		o := WithTaskTools(conf.TaskTools[plan.Name])
		o(taskLLM)
		taskLLM.LLMClient.GetUserMessage(plan.Description)
		taskLLM.Content = plan.Description
		
		logger.Info("execute task", "task", plan.Name)
		err := d.requestTask(ctx, taskLLM, plan)
		if err != nil {
			return err
		}
		d.Token += taskLLM.Token
		completeTasks[plan.Description] = true
	}
	
	llm.LLMClient.AppendMessages(taskLLM.LLMClient)
	
	taskParam := map[string]interface{}{
		"user_task":      d.Content,
		"complete_tasks": completeTasks,
		"last_plan":      lastPlan,
	}
	
	llm.LLMClient.GetUserMessage(i18n.GetMessage(*conf.BaseConfInfo.Lang, "loop_task_prompt", taskParam))
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return err
	}
	
	if len(c) == 0 {
		logger.Error("response is emtpy", "response", c)
		return errors.New("response is emtpy")
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
		return nil
	}
	
	return d.loopTask(ctx, plans, c, llm, loop+1)
}

// requestTask request task
func (d *DeepseekTaskReq) requestTask(ctx context.Context, llm *LLM, plan *Task) error {
	
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return err
	}
	
	// deepseek response merge into msg
	if c == "" {
		c = plan.Name + " is completed"
	}
	llm.LLMClient.GetAssistantMessage(c)
	
	return nil
}
