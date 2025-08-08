package llm

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
)

var (
	jsonRe = regexp.MustCompile(`(?s)\{[\s\r\n]*"plan"\s*:\s*\[.*?][\s\r\n]*}`)
)

type LLMTaskReq struct {
	MessageChan chan *param.MsgInfo
	HTTPMsgChan chan string
	Content     string
	Model       string
	Token       int
	
	UserId string
	ChatId string
	MsgId  string
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
func (d *LLMTaskReq) ExecuteTask() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	
	logger.Info("task content", "content", d.Content)
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
	
	prompt := i18n.GetMessage(*conf.BaseConfInfo.Lang, "assign_task_prompt", taskParam)
	llm := NewLLM(WithUserId(d.UserId), WithChatId(d.ChatId), WithMsgId(d.MsgId),
		WithMessageChan(d.MessageChan), WithContent(prompt), WithHTTPMsgChan(d.HTTPMsgChan))
	llm.LLMClient.GetUserMessage(prompt)
	llm.LLMClient.GetModel(llm)
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("get message fail", "err", err)
		return err
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
		
		finalLLM := NewLLM(WithUserId(d.UserId), WithChatId(d.ChatId), WithMsgId(d.MsgId),
			WithMessageChan(d.MessageChan), WithContent(d.Content))
		finalLLM.LLMClient.GetUserMessage(c)
		finalLLM.LLMClient.GetModel(finalLLM)
		err = finalLLM.LLMClient.Send(ctx, finalLLM)
		if err != nil {
			logger.Error("request summary fail", "err", err)
		}
		return err
	}
	
	llm.LLMClient.GetAssistantMessage(c)
	err = d.loopTask(ctx, plans, c, llm, 0)
	if err != nil {
		logger.Error("loopTask fail", "err", err)
		return err
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
	}
	
	return err
}

// loopTask loop task
func (d *LLMTaskReq) loopTask(ctx context.Context, plans *TaskInfo, lastPlan string, llm *LLM, loop int) error {
	if loop > MostLoop {
		return errors.New("too many loops")
	}
	
	completeTasks := map[string]bool{}
	taskLLM := NewLLM(WithUserId(d.UserId), WithChatId(d.ChatId), WithMsgId(d.MsgId),
		WithMessageChan(d.MessageChan))
	for _, plan := range plans.Plan {
		toolInter, ok := conf.TaskTools.Load(plan.Name)
		var tool *conf.AgentInfo
		if ok {
			tool = toolInter.(*conf.AgentInfo)
		}
		WithTaskTools(tool)(taskLLM)
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
func (d *LLMTaskReq) requestTask(ctx context.Context, llm *LLM, plan *Task) error {
	
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return err
	}
	
	// llm response merge into msg
	if c == "" {
		c = plan.Name + " is completed"
	}
	llm.LLMClient.GetAssistantMessage(c)
	
	return nil
}
