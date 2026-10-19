package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
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

	Cs  *param.ContextState
	Ctx context.Context

	// ConfigExecutor runs MuseBot configuration actions requested inside the
	// task workflow; injected by the robot layer.
	ConfigExecutor func(action *param.ConversationConfigAction) (string, error)
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

	machine := newTaskStateMachine()
	if err := machine.beginResearch(); err != nil {
		return err
	}
	if err := d.runResearch(machine); err != nil {
		logger.ErrorCtx(d.Ctx, "task research fail", "err", err)
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

	if err := machine.transition(TaskStateSummarize); err != nil {
		return err
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

func (d *LLMTaskReq) runResearch(machine *taskStateMachine) error {
	if err := machine.requireState(TaskStateResearchQuery); err != nil {
		return err
	}
	for machine.state == TaskStateResearchQuery {
		taskParam := map[string]interface{}{
			"user_task": d.Content,
			"round":     machine.researchRound + 1,
			"evidence":  strings.Join(machine.researchEvidence, "\n\n"),
		}
		prompt := i18n.GetMessage("task_research_prompt", taskParam)
		llm := d.newTaskLLM(prompt, nil, false)
		content, err := d.syncTaskLLM(llm)
		if err != nil {
			logger.WarnCtx(d.Ctx, "task research query failed; using fallback query", "err", err)
			content = ""
		}

		queries := normalizeTaskResearchPlan(content, d.fallbackSearchQuery())
		d.sendProgress(i18n.GetMessage("task_progress_research", map[string]interface{}{
			"round":   machine.researchRound + 1,
			"queries": strings.Join(queries, "; "),
		}))
		if err := machine.transition(TaskStateResearchObserve); err != nil {
			return err
		}

		for _, query := range queries {
			response, searchErr := utils.WebSearch(d.Ctx, query)
			if searchErr != nil {
				machine.addResearchEvidence(fmt.Sprintf("web_search(%q) failed: %v", query, searchErr))
				continue
			}
			machine.addResearchEvidence(formatWebSearchEvidence(query, response))
		}

		if err := machine.transition(TaskStateResearchReview); err != nil {
			return err
		}
		machine.researchRound++
		if err := machine.beginNextResearchRound(); err != nil {
			return err
		}
	}
	return nil
}

func (d *LLMTaskReq) fallbackSearchQuery() string {
	query := strings.TrimSpace(d.Content)
	if utf8.RuneCountInString(query) <= taskMaxQueryRunes {
		return query
	}
	runes := []rune(query)
	return string(runes[:taskMaxQueryRunes])
}

func (m *taskStateMachine) addResearchEvidence(evidence string) {
	evidence = strings.TrimSpace(evidence)
	if evidence == "" {
		return
	}
	m.researchEvidence = append(m.researchEvidence, evidence)

	total := 0
	kept := make([]string, 0, len(m.researchEvidence))
	for index := len(m.researchEvidence) - 1; index >= 0; index-- {
		item := m.researchEvidence[index]
		if total+len(item) > taskMaxObservationLen {
			break
		}
		total += len(item)
		kept = append([]string{item}, kept...)
	}
	m.researchEvidence = kept
}

func formatWebSearchEvidence(query, response string) string {
	var results utils.WebSearchResponse
	if err := json.Unmarshal([]byte(response), &results); err != nil {
		return fmt.Sprintf("web_search(%q):\n%s", query, response)
	}

	var output strings.Builder
	fmt.Fprintf(&output, "web_search(%q):", query)
	for _, result := range results.Results {
		fmt.Fprintf(&output, "\n- %s\n  %s", result.Title, result.URL)
		if strings.TrimSpace(result.Snippet) != "" {
			fmt.Fprintf(&output, "\n  %s", result.Snippet)
		}
	}
	return output.String()
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
			"evidence":     strings.Join(machine.researchEvidence, "\n\n"),
			"assign_param": d.taskAgentParams(),
		}
		prompt := i18n.GetMessage("task_state_plan_prompt", taskParam)
		// the planner is reused by summarizeTask, which streams its answer to the
		// user, so it must be created with the parent's channels attached
		llm := d.newTaskLLM(prompt, nil, d.MessageChan != nil || d.HTTPMsgChan != nil)
		content, err := d.syncTaskLLM(llm)
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
			"evidence":         strings.Join(machine.researchEvidence, "\n\n"),
			"completed_tasks":  formatTaskResults(machine.completedResults),
		}
		prompt := i18n.GetMessage("task_state_execute_prompt", taskParam)
		if tool != nil && tool.Instructions != "" {
			prompt = tool.Instructions + "\n\n" + prompt
		}

		llm := d.newTaskLLM(prompt, tool, true)
		result, err := d.syncTaskLLM(llm)
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
		"evidence":        strings.Join(machine.researchEvidence, "\n\n"),
		"completed_tasks": formatTaskResults(machine.completedResults),
		"execution_loop":  machine.executionLoop,
	}
	prompt := i18n.GetMessage("task_state_reflect_prompt", taskParam)
	d.sendProgress(i18n.GetMessage("task_progress_reflect", nil))
	planner.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	planner.Content = prompt
	planner.LLMClient.GetModel(planner)

	content, err := d.syncTaskLLM(planner)
	if err != nil {
		return nil, err
	}
	plan := new(TaskInfo)
	if err := extractTaskJSON(content, plan); err != nil {
		return nil, err
	}
	if err := validateTaskPlan(plan, false); err != nil {
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
		"evidence":        strings.Join(machine.researchEvidence, "\n\n"),
		"completed_tasks": formatTaskResults(machine.completedResults),
	}
	prompt := i18n.GetMessage("task_state_summary_prompt", taskParam)
	d.sendProgress(i18n.GetMessage("task_progress_summarize", nil))
	planner.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)
	planner.Content = prompt
	planner.LLMClient.GetModel(planner)

	metrics.APIRequestCount.WithLabelValues(planner.Model).Inc()
	return planner.LLMClient.Send(d.Ctx, planner)
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
