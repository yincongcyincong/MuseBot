package llm

import (
	"testing"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestTaskStateMachineStartsAtPlan(t *testing.T) {
	machine := newTaskStateMachine()
	// research is no longer a separate phase: tasks go straight to planning
	if err := machine.transition(TaskStateExecute); err == nil {
		t.Fatal("expected INIT -> EXECUTE to be rejected")
	}
	if err := machine.beginPlan(); err != nil {
		t.Fatalf("beginPlan() error = %v", err)
	}

	transitions := []TaskState{
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

// single-step plans skip reflection: EXECUTE may go straight to SUMMARIZE
func TestTaskStateMachineAllowsDirectSummarize(t *testing.T) {
	machine := newTaskStateMachine()
	if err := machine.beginPlan(); err != nil {
		t.Fatalf("beginPlan() error = %v", err)
	}
	if err := machine.beginExecutionLoop(); err != nil {
		t.Fatalf("beginExecutionLoop() error = %v", err)
	}
	if err := machine.transition(TaskStateSummarize); err != nil {
		t.Fatalf("EXECUTE -> SUMMARIZE should be allowed: %v", err)
	}
	if err := machine.transition(TaskStateDone); err != nil {
		t.Fatalf("SUMMARIZE -> DONE failed: %v", err)
	}
}

func TestTaskStateMachineLimitsExecutionLoops(t *testing.T) {
	machine := newTaskStateMachine()
	if err := machine.beginPlan(); err != nil {
		t.Fatalf("beginPlan() error = %v", err)
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

func TestTaskStateMachineUsesConfiguredLoopLimit(t *testing.T) {
	previous := conf.BaseConfInfo.TaskMaxExecLoops
	defer func() { conf.BaseConfInfo.TaskMaxExecLoops = previous }()

	conf.BaseConfInfo.TaskMaxExecLoops = 5
	if got := newTaskStateMachine().maxExecutionLoops; got != 5 {
		t.Fatalf("maxExecutionLoops = %d, want 5", got)
	}

	// unset or invalid values fall back to the built-in default
	for _, invalid := range []int{0, -1} {
		conf.BaseConfInfo.TaskMaxExecLoops = invalid
		if got := newTaskStateMachine().maxExecutionLoops; got != taskMaxExecutionLoops {
			t.Fatalf("fallback for %d = %d, want %d", invalid, got, taskMaxExecutionLoops)
		}
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
