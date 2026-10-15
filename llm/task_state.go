package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	taskMaxResearchRounds  = 3
	taskMaxQueriesPerRound = 2
	taskMaxPlanItems       = 8
	taskMaxExecutionLoops  = 3
	taskMaxQueryRunes      = 256
	taskMaxObservationLen  = 12 * 1024
)

type TaskState string

const (
	TaskStateInit            TaskState = "INIT"
	TaskStateResearchQuery   TaskState = "RESEARCH_QUERY"
	TaskStateResearchObserve TaskState = "RESEARCH_OBSERVE"
	TaskStateResearchReview  TaskState = "RESEARCH_REVIEW"
	TaskStatePlan            TaskState = "PLAN"
	TaskStateExecute         TaskState = "EXECUTE"
	TaskStateReflect         TaskState = "REFLECT"
	TaskStateSummarize       TaskState = "SUMMARIZE"
	TaskStateDone            TaskState = "DONE"
)

var taskStateTransitions = map[TaskState][]TaskState{
	TaskStateInit:            {TaskStateResearchQuery},
	TaskStateResearchQuery:   {TaskStateResearchObserve},
	TaskStateResearchObserve: {TaskStateResearchReview},
	TaskStateResearchReview:  {TaskStateResearchQuery, TaskStatePlan},
	TaskStatePlan:            {TaskStateExecute},
	TaskStateExecute:         {TaskStateReflect},
	TaskStateReflect:         {TaskStatePlan, TaskStateSummarize},
	TaskStateSummarize:       {TaskStateDone},
}

type taskStateMachine struct {
	state            TaskState
	researchRound    int
	executionLoop    int
	researchEvidence []string
	completedResults []TaskResult
}

type taskResearchPlan struct {
	Queries []string `json:"queries"`
}

func newTaskStateMachine() *taskStateMachine {
	return &taskStateMachine{
		state: TaskStateInit,
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

func (m *taskStateMachine) beginResearch() error {
	return m.transition(TaskStateResearchQuery)
}

func (m *taskStateMachine) beginNextResearchRound() error {
	if m.researchRound >= taskMaxResearchRounds {
		return m.transition(TaskStatePlan)
	}
	return m.transition(TaskStateResearchQuery)
}

func (m *taskStateMachine) beginExecutionLoop() error {
	if m.executionLoop >= taskMaxExecutionLoops {
		return fmt.Errorf("task exceeded %d execution loops", taskMaxExecutionLoops)
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

func normalizeTaskResearchPlan(content string, fallbackQuery string) []string {
	var research taskResearchPlan
	if err := extractTaskJSON(content, &research); err != nil {
		return []string{fallbackQuery}
	}

	seen := make(map[string]bool)
	queries := make([]string, 0, taskMaxQueriesPerRound)
	for _, query := range research.Queries {
		query = strings.TrimSpace(query)
		if query == "" || utf8.RuneCountInString(query) > taskMaxQueryRunes {
			continue
		}
		key := strings.ToLower(query)
		if seen[key] {
			continue
		}
		seen[key] = true
		queries = append(queries, query)
		if len(queries) >= taskMaxQueriesPerRound {
			break
		}
	}

	if len(queries) == 0 {
		return []string{fallbackQuery}
	}
	return queries
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
	output.WriteString("Task state: PLAN\n")
	for index, task := range plan.Plan {
		fmt.Fprintf(&output, "%d. [%s] %s\n", index+1, task.Name, task.Description)
	}
	return strings.TrimRight(output.String(), "\n")
}
