package llm

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

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
	PerMsgLen   int

	UserId string
	ChatId string
	MsgId  string

	Ctx context.Context
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
	logger.InfoCtx(d.Ctx, "task content", "content", d.Content)
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
		WithMessageChan(d.MessageChan), WithContent(prompt), WithHTTPMsgChan(d.HTTPMsgChan),
		WithPerMsgLen(d.PerMsgLen), WithContext(d.Ctx))
	llm.LLMClient.GetUserMessage(prompt)
	llm.LLMClient.GetModel(llm)
	c, err := llm.LLMClient.SyncSend(d.Ctx, llm)
	if err != nil {
		logger.ErrorCtx(d.Ctx, "get message fail", "err", err)
		return err
	}

	d.Token += llm.Token

	matches := jsonRe.FindAllString(c, -1)
	plans := new(TaskInfo)
	for _, match := range matches {
		err = json.Unmarshal([]byte(match), &plans)
		if err != nil {
			logger.WarnCtx(d.Ctx, "json umarshal fail", "err", err)
		}
	}

	logger.InfoCtx(d.Ctx, "task plan", "plan", plans)

	if len(plans.Plan) == 0 {
		logger.InfoCtx(d.Ctx, "no plan created!")

		finalLLM := NewLLM(WithUserId(d.UserId), WithChatId(d.ChatId), WithMsgId(d.MsgId),
			WithMessageChan(d.MessageChan), WithContent(d.Content), WithHTTPMsgChan(d.HTTPMsgChan),
			WithPerMsgLen(d.PerMsgLen), WithContext(d.Ctx))
		finalLLM.LLMClient.GetUserMessage(c)
		finalLLM.LLMClient.GetModel(finalLLM)
		err = finalLLM.LLMClient.Send(d.Ctx, finalLLM)
		if err != nil {
			logger.ErrorCtx(d.Ctx, "request summary fail", "err", err)
		}
		return err
	}

	llm.DirectSendMsg(c)
	llm.LLMClient.GetAssistantMessage(c)
	err = d.loopTask(d.Ctx, plans, c, llm, 0)
	if err != nil {
		logger.ErrorCtx(d.Ctx, "loopTask fail", "err", err)
		return err
	}

	// summary
	summaryParam := make(map[string]interface{})
	summaryParam["user_task"] = d.Content
	summaryPrompt := i18n.GetMessage(*conf.BaseConfInfo.Lang, "summary_task_prompt", summaryParam)
	llm.LLMClient.GetUserMessage(summaryPrompt)
	llm.Content = summaryPrompt
	err = llm.LLMClient.Send(d.Ctx, llm)
	if err != nil {
		logger.ErrorCtx(d.Ctx, "request summary fail", "err", err)
		return err
	}

	err = llm.InsertOrUpdate()
	if err != nil {
		logger.ErrorCtx(d.Ctx, "insertOrUpdate fail", "err", err)
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
		WithMessageChan(d.MessageChan), WithHTTPMsgChan(d.HTTPMsgChan), WithPerMsgLen(d.PerMsgLen),
		WithContext(d.Ctx))
	for _, plan := range plans.Plan {
		toolInter, ok := conf.TaskTools.Load(plan.Name)
		var tool *conf.AgentInfo
		if ok {
			tool = toolInter.(*conf.AgentInfo)
		}
		WithTaskTools(tool)(taskLLM)
		taskLLM.LLMClient.GetUserMessage(plan.Description)
		taskLLM.Content = plan.Description
		taskLLM.LLMClient.GetModel(taskLLM)
		logger.InfoCtx(d.Ctx, "execute task", "task", plan.Name, "task desc", plan.Description)
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
	llm.LLMClient.GetModel(llm)
	c, err := llm.LLMClient.SyncSend(ctx, llm)
	if err != nil {
		logger.ErrorCtx(d.Ctx, "ChatCompletionStream error", "err", err)
		return err
	}

	if len(c) == 0 {
		logger.ErrorCtx(d.Ctx, "response is emtpy", "response", c)
		return errors.New("response is emtpy")
	}

	d.Token += llm.Token

	matches := jsonRe.FindAllString(c, -1)
	plans = new(TaskInfo)
	for _, match := range matches {
		err := json.Unmarshal([]byte(match), &plans)
		if err != nil {
			logger.ErrorCtx(d.Ctx, "json umarshal fail", "err", err)
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
		logger.ErrorCtx(d.Ctx, "ChatCompletionStream error", "err", err)
		return err
	}

	// llm response merge into msg
	c += "\n\n" + plan.Name + " is completed"
	llm.LLMClient.GetAssistantMessage(c)

	return nil
}
