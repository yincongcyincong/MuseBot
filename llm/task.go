package llm

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type LLMTaskReq struct {
	// One identity per standalone research run, shared by all of its agents.
	searchSession utils.WebSearchSession

	MessageChan chan *param.MsgInfo
	HTTPMsgChan chan string
	Content     string
	Model       string
	Token       int
	PerMsgLen   int

	UserId string
	ChatId string
	MsgId  string

	Cs  *param.ContextState
	Ctx context.Context

	// ConfigExecutor runs MuseBot configuration actions requested inside the
	// task workflow; injected by the robot layer.
	ConfigExecutor func(action *param.ConversationConfigAction, approve TaskApprovalFunc) (string, error)
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

func (d *LLMTaskReq) ExecuteTask() error {
	logger.InfoCtx(d.Ctx, "task content", "content", d.Content)

	// research is not a separate phase: the planner decides whether a web
	// search step is needed and assigns the web_search agent like any other
	machine := newTaskStateMachine()
	if err := machine.beginPlan(); err != nil {
		return err
	}

	planner, plan, err := d.planTask(machine)
	if err != nil {
		return err
	}

	for machine.state == TaskStatePlan {
		if err := machine.beginExecutionLoop(); err != nil {
			return err
		}
		if err := d.executePlan(machine, plan); err != nil {
			return err
		}

		// a single-step plan that completed needs no reflection round: there
		// is nothing left to re-evaluate, so summarize directly
		if len(plan.Plan) == 1 {
			if err := machine.transition(TaskStateSummarize); err != nil {
				return err
			}
			break
		}

		if err := machine.transition(TaskStateReflect); err != nil {
			return err
		}
		plan, err = d.reflectTask(planner, machine)
		if err != nil {
			return err
		}
		if len(plan.Plan) == 0 {
			break
		}

		if err := machine.transition(TaskStatePlan); err != nil {
			return err
		}
		planOutput := formatTaskPlan(plan)
		planner.DirectSendMsg(planOutput, false)
		planner.LLMClient.GetMessage(openai.ChatMessageRoleAssistant, planOutput)
	}

	if machine.state != TaskStateSummarize {
		if err := machine.transition(TaskStateSummarize); err != nil {
			return err
		}
	}
	if err := d.summarizeTask(planner, machine); err != nil {
		return err
	}
	if err := machine.finish(); err != nil {
		return err
	}
	return planner.InsertOrUpdate()
}

func (d *LLMTaskReq) taskAgentParams() []map[string]string {
	agents := make([]map[string]string, 0)
	conf.TaskTools.Range(func(name, value any) bool {
		tool := value.(*conf.AgentInfo)
		agents = append(agents, map[string]string{
			"tool_name": name.(string),
			"tool_desc": tool.Description,
		})
		return true
	})
	sort.Slice(agents, func(left, right int) bool {
		return agents[left]["tool_name"] < agents[right]["tool_name"]
	})
	return agents
}

func (d *LLMTaskReq) newTaskLLM(prompt string, tool *conf.AgentInfo, streaming bool) *LLM {
	options := []Option{
		WithUserId(d.UserId),
		WithChatId(d.ChatId),
		WithMsgId(d.MsgId),
		WithContent(prompt),
		WithPerMsgLen(d.PerMsgLen),
		WithContext(d.Ctx),
		WithCS(d.Cs),
		WithTaskTools(tool),
		WithToolApprovals(true),
	}
	if streaming {
		options = append(options, WithMessageChan(d.MessageChan), WithHTTPMsgChan(d.HTTPMsgChan))
	}
	if d.ConfigExecutor != nil {
		options = append(options, WithConfigExecutor(d.ConfigExecutor))
	}

	llm := NewLLM(options...)
	llm.searchSession = &d.searchSession
	if tool == nil {
		// the planner (and fallback agents without a tool) sees recent chat
		// history so references like "the cron task I created earlier"
		// resolve against previous conversations
		llm.appendHistory(d.UserId)
	}
	llm.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	llm.LLMClient.GetModel(llm)
	return llm
}

func (d *LLMTaskReq) syncTaskLLM(llm *LLM) (string, error) {
	metrics.APIRequestCount.WithLabelValues(llm.Model).Inc()
	content, err := llm.LLMClient.SyncSend(d.Ctx, llm)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("task state returned an empty response")
	}
	return content, nil
}

// retryTaskRequest reruns a task-stage LLM request when it fails, so a
// transient API error does not abort the whole task run.
func (d *LLMTaskReq) retryTaskRequest(stage string, fn func() error) error {
	retryTimes := conf.BaseConfInfo.LLMRetryTimes
	if retryTimes <= 0 {
		retryTimes = 1
	}
	var err error
	for attempt := 0; attempt < retryTimes; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		logger.WarnCtx(d.Ctx, "task llm request fail, retrying", "stage", stage, "attempt", attempt+1, "err", err)
	}
	return err
}

// sendProgress pushes a short status message to the user's chat while the
// task workflow advances between stages, so long-running tasks stay responsive.
func (d *LLMTaskReq) sendProgress(message string) {
	if message == "" {
		return
	}
	perMsgLen := d.PerMsgLen
	if perMsgLen == 0 {
		perMsgLen = OneMsgLen
	}
	if len([]byte(message)) > perMsgLen {
		message = string([]byte(message)[:perMsgLen])
	}

	if d.MessageChan != nil {
		d.MessageChan <- &param.MsgInfo{
			Content:  message,
			Finished: true,
		}
	}
	if d.HTTPMsgChan != nil {
		d.HTTPMsgChan <- message
	}
}

func (d *LLMTaskReq) planTask(machine *taskStateMachine) (*LLM, *TaskInfo, error) {
	if err := machine.requireState(TaskStatePlan); err != nil {
		return nil, nil, err
	}
	var lastErr error
	d.sendProgress(i18n.GetMessage("task_progress_plan", nil))
	for attempt := 0; attempt < 2; attempt++ {
		taskParam := map[string]interface{}{
			"user_task":    d.Content,
			"assign_param": d.taskAgentParams(),
		}
		prompt := i18n.GetMessage("task_state_plan_prompt", taskParam)
		// the planner is reused by summarizeTask, which streams its answer to the
		// user, so it must be created with the parent's channels attached
		llm := d.newTaskLLM(prompt, nil, d.MessageChan != nil || d.HTTPMsgChan != nil)
		var content string
		err := d.retryTaskRequest("plan", func() error {
			var syncErr error
			content, syncErr = d.syncTaskLLM(llm)
			return syncErr
		})
		if err != nil {
			return nil, nil, err
		}

		plan := new(TaskInfo)
		if err := extractTaskJSON(content, plan); err == nil {
			err = validateTaskPlan(plan, true)
		}
		if err == nil {
			planOutput := formatTaskPlan(plan)
			llm.DirectSendMsg(planOutput, false)
			llm.LLMClient.GetMessage(openai.ChatMessageRoleAssistant, planOutput)
			return llm, plan, nil
		}
		lastErr = err
		logger.WarnCtx(d.Ctx, "invalid task plan", "attempt", attempt+1, "err", err)
	}
	return nil, nil, fmt.Errorf("invalid task plan: %w", lastErr)
}

func (d *LLMTaskReq) executePlan(machine *taskStateMachine, plan *TaskInfo) error {
	if err := machine.requireState(TaskStateExecute); err != nil {
		return err
	}
	for i, task := range plan.Plan {
		toolValue, loaded := conf.TaskTools.Load(task.Name)
		var tool *conf.AgentInfo
		if loaded {
			tool = toolValue.(*conf.AgentInfo)
		}

		d.sendProgress(i18n.GetMessage("task_progress_execute", map[string]interface{}{
			"index":       i + 1,
			"total":       len(plan.Plan),
			"name":        task.Name,
			"description": task.Description,
		}))
		taskParam := map[string]interface{}{
			"user_task":        d.Content,
			"agent_name":       task.Name,
			"task_description": task.Description,
			"completed_tasks":  formatTaskResults(machine.completedResults),
		}
		prompt := i18n.GetMessage("task_state_execute_prompt", taskParam)
		if tool != nil && tool.Instructions != "" {
			prompt = tool.Instructions + "\n\n" + prompt
		}

		llm := d.newTaskLLM(prompt, tool, true)
		var result string
		err := d.retryTaskRequest("execute", func() error {
			var syncErr error
			result, syncErr = d.syncTaskLLM(llm)
			return syncErr
		})
		if err != nil {
			return err
		}
		d.Token += llm.Cs.Token
		d.sendProgress(i18n.GetMessage("task_progress_execute_done", map[string]interface{}{
			"index": i + 1,
			"total": len(plan.Plan),
			"name":  task.Name,
		}))
		machine.completedResults = append(machine.completedResults, TaskResult{
			TaskName:   task.Name,
			TaskResult: result,
		})
	}
	return nil
}

func (d *LLMTaskReq) reflectTask(planner *LLM, machine *taskStateMachine) (*TaskInfo, error) {
	if err := machine.requireState(TaskStateReflect); err != nil {
		return nil, err
	}
	taskParam := map[string]interface{}{
		"user_task":       d.Content,
		"completed_tasks": formatTaskResults(machine.completedResults),
		"execution_loop":  machine.executionLoop,
	}
	prompt := i18n.GetMessage("task_state_reflect_prompt", taskParam)
	d.sendProgress(i18n.GetMessage("task_progress_reflect", nil))
	planner.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	planner.Content = prompt
	planner.LLMClient.GetModel(planner)

	var content string
	var plan *TaskInfo
	err := d.retryTaskRequest("reflect", func() error {
		var syncErr error
		content, syncErr = d.syncTaskLLM(planner)
		if syncErr != nil {
			return syncErr
		}
		plan = new(TaskInfo)
		if err := extractTaskJSON(content, plan); err != nil {
			return err
		}
		return validateTaskPlan(plan, false)
	})
	if err != nil {
		return nil, err
	}
	if len(plan.Plan) > 0 {
		planner.LLMClient.GetMessage(openai.ChatMessageRoleAssistant, formatTaskPlan(plan))
	} else {
		planner.LLMClient.GetMessage(openai.ChatMessageRoleAssistant, content)
	}
	return plan, nil
}

func (d *LLMTaskReq) summarizeTask(planner *LLM, machine *taskStateMachine) error {
	if err := machine.requireState(TaskStateSummarize); err != nil {
		return err
	}
	taskParam := map[string]interface{}{
		"user_task":       d.Content,
		"completed_tasks": formatTaskResults(machine.completedResults),
	}
	prompt := i18n.GetMessage("task_state_summary_prompt", taskParam)
	d.sendProgress(i18n.GetMessage("task_progress_summarize", nil))
	planner.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	planner.Content = prompt
	planner.LLMClient.GetModel(planner)

	metrics.APIRequestCount.WithLabelValues(planner.Model).Inc()
	return d.retryTaskRequest("summarize", func() error {
		return planner.LLMClient.Send(d.Ctx, planner)
	})
}

func formatTaskResults(results []TaskResult) string {
	if len(results) == 0 {
		return "(none)"
	}
	var output strings.Builder
	for _, result := range results {
		fmt.Fprintf(&output, "- [%s]\n%s\n", result.TaskName, result.TaskResult)
	}
	return strings.TrimRight(output.String(), "\n")
}
