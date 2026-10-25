package llm

import (
	"encoding/json"

	"errors"
	"fmt"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/i18n"
	"strings"
)

const (
	taskMaxPlanItems = 8
	// taskMaxExecutionLoops is the fallback used when the configured value is
	// unset or invalid, so existing config files keep the previous behaviour.
	taskMaxExecutionLoops = 3
)

type TaskState string

const (
	TaskStateInit      TaskState = "INIT"
	TaskStatePlan      TaskState = "PLAN"
	TaskStateExecute   TaskState = "EXECUTE"
	TaskStateReflect   TaskState = "REFLECT"
	TaskStateSummarize TaskState = "SUMMARIZE"
	TaskStateDone      TaskState = "DONE"
)

var taskStateTransitions = map[TaskState][]TaskState{
	TaskStateInit:      {TaskStatePlan},
	TaskStatePlan:      {TaskStateExecute},
	TaskStateExecute:   {TaskStateReflect, TaskStateSummarize},
	TaskStateReflect:   {TaskStatePlan, TaskStateSummarize},
	TaskStateSummarize: {TaskStateDone},
}

type taskStateMachine struct {
	state             TaskState
	executionLoop     int
	maxExecutionLoops int
	completedResults  []TaskResult
}

func newTaskStateMachine() *taskStateMachine {
	maxLoops := conf.BaseConfInfo.TaskMaxExecLoops
	if maxLoops <= 0 {
		maxLoops = taskMaxExecutionLoops
	}
	return &taskStateMachine{
		state:             TaskStateInit,
		maxExecutionLoops: maxLoops,
	}
}

func (m *taskStateMachine) transition(next TaskState) error {
	if m.state == TaskStateDone {
		return errors.New("task state machine is done")
	}
	for _, allowed := range taskStateTransitions[m.state] {
		if next == allowed {
			m.state = next
			return nil
		}
	}

	return fmt.Errorf("invalid task state transition %s -> %s", m.state, next)
}

func (m *taskStateMachine) requireState(expected TaskState) error {
	if m.state != expected {
		return fmt.Errorf("task state %s does not match required state %s", m.state, expected)
	}
	return nil
}

func (m *taskStateMachine) beginPlan() error {
	return m.transition(TaskStatePlan)
}

func (m *taskStateMachine) beginExecutionLoop() error {
	if m.executionLoop >= m.maxExecutionLoops {
		return fmt.Errorf("task exceeded %d execution loops", m.maxExecutionLoops)
	}
	if err := m.transition(TaskStateExecute); err != nil {
		return err
	}
	m.executionLoop++
	return nil
}

func (m *taskStateMachine) finish() error {
	return m.transition(TaskStateDone)
}

func validateTaskPlan(plan *TaskInfo, requireNonEmpty bool) error {
	if plan == nil {
		if requireNonEmpty {
			return errors.New("task plan is required")
		}
		return nil
	}
	if len(plan.Plan) > taskMaxPlanItems {
		return fmt.Errorf("task plan has more than %d items", taskMaxPlanItems)
	}
	if requireNonEmpty && len(plan.Plan) == 0 {
		return errors.New("task plan is required")
	}

	seen := make(map[string]bool)
	for index, task := range plan.Plan {
		if task == nil || strings.TrimSpace(task.Name) == "" || strings.TrimSpace(task.Description) == "" {
			return fmt.Errorf("task plan item %d requires name and description", index+1)
		}
		task.Name = strings.TrimSpace(task.Name)
		task.Description = strings.TrimSpace(task.Description)
		key := task.Name + "\x00" + task.Description
		if seen[key] {
			return fmt.Errorf("task plan item %d duplicates another item", index+1)
		}
		seen[key] = true
	}
	return nil
}

func extractTaskJSON(content string, target any) error {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), target); err == nil {
		return nil
	}

	start := strings.IndexByte(content, '{')
	if start < 0 {
		return errors.New("task response does not contain JSON")
	}
	depth := 0
	inString := false
	escaped := false
	for index := start; index < len(content); index++ {
		character := content[index]
		if inString {
			if escaped {
				escaped = false
			} else if character == '\\' {
				escaped = true
			} else if character == '"' {
				inString = false
			}
			continue
		}
		switch character {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return json.Unmarshal([]byte(content[start:index+1]), target)
			}
		}
	}
	return errors.New("task response contains incomplete JSON")
}

func formatTaskPlan(plan *TaskInfo) string {
	var output strings.Builder
	output.WriteString(i18n.GetMessage("task_plan_header", nil) + "\n")
	for index, task := range plan.Plan {
		fmt.Fprintf(&output, "%d. [%s] %s\n", index+1, task.Name, task.Description)
	}
	return strings.TrimRight(output.String(), "\n")
}
