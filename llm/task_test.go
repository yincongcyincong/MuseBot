package llm

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
)

// The task planner must see recent chat history so follow-up tasks like
// "delete the cron task I created earlier" can resolve against previous
// conversations; execute agents stay focused on their step.
func TestNewTaskLLM_PlannerSeesHistory(t *testing.T) {
	userId := "task-history-user"
	db.MsgRecord.Store(userId, &db.MsgRecordInfo{
		AQs: []*db.AQ{{
			Question:   "创建一个定时任务",
			Answer:     "created cron task 7 (cron: 0 9 * * *, prompt: daily report)",
			CreateTime: time.Now().Add(10 * time.Minute).Unix(),
		}},
	})
	t.Cleanup(func() { db.MsgRecord.Delete(userId) })

	d := &LLMTaskReq{
		UserId: userId,
		Cs:     &param.ContextState{UseRecord: true},
		Ctx: context.WithValue(context.Background(), "user_info", &db.User{
			LLMConfigRaw: &param.LLMConfig{TxtType: param.OpenAi},
		}),
	}

	planner := d.newTaskLLM("plan prompt", nil, false)
	plannerMsgs := planner.LLMClient.(*OpenAIReq).OpenAIMsgs
	assert.GreaterOrEqual(t, len(plannerMsgs), 3)
	assert.Equal(t, "创建一个定时任务", plannerMsgs[0].Content)
	assert.Equal(t, "created cron task 7 (cron: 0 9 * * *, prompt: daily report)", plannerMsgs[1].Content)
	assert.Equal(t, "plan prompt", plannerMsgs[len(plannerMsgs)-1].Content)

	agent := d.newTaskLLM("agent prompt", &conf.AgentInfo{}, false)
	agentMsgs := agent.LLMClient.(*OpenAIReq).OpenAIMsgs
	assert.Equal(t, 1, len(agentMsgs))
	assert.Equal(t, "agent prompt", agentMsgs[0].Content)
}

func TestNewTaskLLM_NoHistoryWithoutRecords(t *testing.T) {
	d := &LLMTaskReq{
		UserId: "task-no-history-user",
		Cs:     &param.ContextState{UseRecord: true},
		Ctx: context.WithValue(context.Background(), "user_info", &db.User{
			LLMConfigRaw: &param.LLMConfig{TxtType: param.OpenAi},
		}),
	}

	planner := d.newTaskLLM("plan prompt", nil, false)
	plannerMsgs := planner.LLMClient.(*OpenAIReq).OpenAIMsgs
	assert.Equal(t, 1, len(plannerMsgs))
	assert.Equal(t, "plan prompt", plannerMsgs[0].Content)
}

// a step counts as failed only when the agent made tool calls and every one
// of them failed; an agent that recovered from one bad call is a success
func TestStepToolFailure(t *testing.T) {
	assert.NoError(t, stepToolFailure(nil))
	assert.NoError(t, stepToolFailure(&LLM{}))                           // no tool calls at all
	assert.NoError(t, stepToolFailure(&LLM{toolCalls: 3, toolFails: 1})) // agent worked around it
	assert.Error(t, stepToolFailure(&LLM{toolCalls: 3, toolFails: 3}))
}

// when a step fails, the remaining steps must not run and the loop must
// surface errStepFailed so executeTask skips reflection and summarizes
func TestExecutePlanStopsAfterFailedStep(t *testing.T) {
	patchConf(t, false, "", 1)
	primary := newFakeProvider(t, http.StatusInternalServerError)
	routeProvider(t, param.OpenAi, primary)

	d := &LLMTaskReq{
		UserId: "task-fail-user",
		Cs:     &param.ContextState{},
		Ctx: context.WithValue(context.Background(), "user_info", &db.User{
			LLMConfigRaw: &param.LLMConfig{TxtType: param.OpenAi, TxtModel: "m"},
		}),
	}

	machine := newTaskStateMachine()
	assert.NoError(t, machine.beginPlan())
	assert.NoError(t, machine.beginExecutionLoop())
	plan := &TaskInfo{Plan: []*Task{
		{Name: "web_search", Description: "first"},
		{Name: "web_search", Description: "second"},
	}}

	err := d.executePlan(machine, plan)
	assert.ErrorIs(t, err, errStepFailed)
	// only the first step was sent to the LLM; the second never started
	assert.Equal(t, 1, primary.hits)
	assert.Len(t, machine.completedResults, 1)
	assert.Contains(t, machine.completedResults[0].TaskResult, "(step failed)")
}
