package taskflow

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
)

func TestMain(m *testing.M) {
	conf.InitConf()
	db.InitTable()
	os.Exit(m.Run())
}

// 未注入跟踪（非 /task 场景）时，所有方法都应是零开销空操作
func TestNilTrackerIsNoop(t *testing.T) {
	var tracker *Tracker

	assert.NotPanics(t, func() {
		tracker.RunStarted()
		tracker.Stage(StateExecute)
		tracker.LoopStarted(1, 3)
		tracker.PlanCreated([]TaskStep{{Name: "s"}})
		tracker.StepStarted(0, "s")
		tracker.StepFinished(0, "s", "", nil)
		tracker.ToolCalled("web_search", "{}", nil, time.Second)
		tracker.ApprovalRequested("write", "/tmp/x")
		tracker.Finish(StateDone, "ok", "")
	})
}

func TestTrackerLifecycle(t *testing.T) {
	tracker := StartRun("研究 Go 并发模型", "user-1", "chat-1")
	assert.NotNil(t, tracker)

	runId := tracker.run.RunId

	// 初始状态：PLAN、出现在运行中列表
	run := GetRun(runId)
	assert.NotNil(t, run)
	assert.Equal(t, StatePlan, run.State)
	assert.Equal(t, "user-1", run.UserId)
	assert.Equal(t, "chat-1", run.ChatId)
	assert.Equal(t, "研究 Go 并发模型", run.Content)
	assert.Positive(t, run.StartedAt)
	assert.True(t, containsRun(RunningRuns(), runId))

	// 执行循环 + 计划
	tracker.RunStarted()
	tracker.Stage(StateExecute)
	tracker.LoopStarted(1, 3)
	tracker.PlanCreated([]TaskStep{
		{Name: "search", Description: "查资料", Status: StepPending},
		{Name: "write", Description: "写总结", Status: StepPending},
	})
	run = GetRun(runId)
	assert.Equal(t, StateExecute, run.State)
	assert.Equal(t, 1, run.Loop)
	assert.Equal(t, 3, run.MaxLoops)
	assert.Len(t, run.Plan, 2)

	// 步骤执行
	tracker.StepStarted(0, "search")
	run = GetRun(runId)
	assert.Equal(t, StepRunning, run.Plan[0].Status)

	tracker.StepFinished(0, "search", "找到了三篇资料", nil)
	tracker.StepFinished(1, "write", "", errors.New("agent failed"))
	run = GetRun(runId)
	assert.Equal(t, StepDone, run.Plan[0].Status)
	assert.Equal(t, "找到了三篇资料", run.Plan[0].Result)
	assert.Equal(t, StepFailed, run.Plan[1].Status)

	// 工具调用与授权事件
	tracker.ToolCalled("web_search", `{"query":"go"}`, nil, 250*time.Millisecond)
	tracker.ApprovalRequested("write", "/tmp/report.md")

	// 终态：DONE + 落库
	tracker.Finish(StateDone, "最终回答内容", "")

	run = GetRun(runId)
	assert.Equal(t, StateDone, run.State)
	assert.Equal(t, "最终回答内容", run.FinalResult)
	assert.Empty(t, run.ErrMsg)
	assert.Positive(t, run.FinishedAt)
	assert.Equal(t, "run_finished", run.Events[len(run.Events)-1].Type)
	assert.False(t, containsRun(RunningRuns(), runId))

	// 落库记录可通过 run_id 取回，steps/events 可解码
	record, err := db.GetTaskRunByRunID(runId)
	assert.NoError(t, err)
	assert.Equal(t, StateDone, record.State)
	assert.Equal(t, "最终回答内容", record.FinalResult)
	assert.Len(t, db.TaskRunSteps(record.Steps), 2)
	assert.NotEmpty(t, db.TaskRunEvents(record.Events))
}

func TestEventsCapped(t *testing.T) {
	tracker := StartRun("压测事件上限", "user-2", "chat-2")
	for i := 0; i < maxEvents+50; i++ {
		tracker.ToolCalled("tool", "{}", nil, time.Millisecond)
	}
	run := GetRun(tracker.run.RunId)
	assert.Len(t, run.Events, maxEvents)
	// 最旧的被丢弃，保留的是最新的事件
	assert.Equal(t, "tool_called", run.Events[0].Type)
}

func containsRun(runs []*TaskRun, runId string) bool {
	for _, run := range runs {
		if run.RunId == runId {
			return true
		}
	}
	return false
}
