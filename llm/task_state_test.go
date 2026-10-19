package llm

import (
	"testing"

	"github.com/yincongcyincong/MuseBot/param"
)

func TestTaskStateMachineEnforcesResearchBeforePlan(t *testing.T) {
	machine := newTaskStateMachine()
	if err := machine.transition(TaskStatePlan); err == nil {
		t.Fatal("expected INIT -> PLAN to be rejected")
	}

	transitions := []TaskState{
		TaskStateResearchQuery,
		TaskStateResearchObserve,
		TaskStateResearchReview,
		TaskStatePlan,
		TaskStateExecute,
		TaskStateReflect,
		TaskStateSummarize,
		TaskStateDone,
	}
	for _, state := range transitions {
		if err := machine.transition(state); err != nil {
			t.Fatalf("transition to %s failed: %v", state, err)
		}
	}
}

func TestTaskStateMachineLimitsResearchAndExecution(t *testing.T) {
	machine := newTaskStateMachine()
	if err := machine.beginResearch(); err != nil {
		t.Fatalf("beginResearch() error = %v", err)
	}

	for round := 0; round < taskMaxResearchRounds; round++ {
		if machine.state != TaskStateResearchQuery {
			t.Fatalf("round %d state = %s, want %s", round, machine.state, TaskStateResearchQuery)
		}
		if err := machine.transition(TaskStateResearchObserve); err != nil {
			t.Fatalf("research observe transition failed: %v", err)
		}
		if err := machine.transition(TaskStateResearchReview); err != nil {
			t.Fatalf("research review transition failed: %v", err)
		}
		machine.researchRound++
		if err := machine.beginNextResearchRound(); err != nil {
			t.Fatalf("beginNextResearchRound() error = %v", err)
		}
	}
	if machine.state != TaskStatePlan {
		t.Fatalf("after research state = %s, want %s", machine.state, TaskStatePlan)
	}

	for loop := 0; loop < taskMaxExecutionLoops; loop++ {
		if machine.state != TaskStatePlan {
			t.Fatalf("execution loop %d state = %s, want %s", loop, machine.state, TaskStatePlan)
		}
		if err := machine.beginExecutionLoop(); err != nil {
			t.Fatalf("beginExecutionLoop() error = %v", err)
		}
		if err := machine.transition(TaskStateReflect); err != nil {
			t.Fatalf("reflect transition failed: %v", err)
		}
		if err := machine.transition(TaskStatePlan); err != nil {
			t.Fatalf("reflect to plan transition failed: %v", err)
		}
	}
	if err := machine.beginExecutionLoop(); err == nil {
		t.Fatal("expected execution loop limit")
	}
}

func TestNormalizeTaskResearchPlan(t *testing.T) {
	queries := normalizeTaskResearchPlan("noise {\"queries\":[\" MCP function calling \",\"mcp function calling\",\"  \"]}", "fallback query")
	if len(queries) != 1 || queries[0] != "MCP function calling" {
		t.Fatalf("unexpected queries: %#v", queries)
	}

	queries = normalizeTaskResearchPlan("not json", "fallback query")
	if len(queries) != 1 || queries[0] != "fallback query" {
		t.Fatalf("fallback queries = %#v", queries)
	}
}

func TestValidateTaskPlan(t *testing.T) {
	if err := validateTaskPlan(&TaskInfo{}, true); err == nil {
		t.Fatal("expected empty initial plan to be rejected")
	}
	if err := validateTaskPlan(&TaskInfo{}, false); err != nil {
		t.Fatalf("empty reflection plan rejected: %v", err)
	}

	plan := &TaskInfo{Plan: []*Task{
		{Name: "web_search", Description: "research"},
		{Name: "writer", Description: "draft"},
		{Name: "writer", Description: "draft"},
	}}
	if err := validateTaskPlan(plan, true); err == nil {
		t.Fatal("expected duplicate plan item to be rejected")
	}
}

func TestSendProgressDeliversToBothChannels(t *testing.T) {
	msgChan := make(chan *param.MsgInfo, 2)
	httpChan := make(chan string, 2)
	d := &LLMTaskReq{
		MessageChan: msgChan,
		HTTPMsgChan: httpChan,
	}

	d.sendProgress("")
	select {
	case <-msgChan:
		t.Error("empty progress should not be sent")
	default:
	}

	d.sendProgress("working...")
	msg := <-msgChan
	if msg.Content != "working..." || !msg.Finished {
		t.Errorf("message chan got %+v", msg)
	}
	if s := <-httpChan; s != "working..." {
		t.Errorf("http chan got %q", s)
	}

	// no channels attached: must not block or panic
	bare := &LLMTaskReq{}
	bare.sendProgress("ignored")
}
