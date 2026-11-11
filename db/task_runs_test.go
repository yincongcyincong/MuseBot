package db

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestTaskRunInsertAndQuery(t *testing.T) {
	runId := "test-run-" + t.Name()

	record := &TaskRunRecord{
		RunId:       runId,
		UserId:      "u1",
		ChatId:      "c1",
		Content:     "写一份周报",
		State:       "DONE",
		LoopCount:   2,
		Steps:       `[{"name":"search","description":"查资料","status":"done","result":"ok"}]`,
		Events:      `[{"ts":1700000000,"type":"run_started","detail":"写一份周报"}]`,
		FinalResult: "周报已生成",
		ErrMsg:      "",
		StartTime:   1700000000,
		EndTime:     1700000100,
	}

	err := InsertTaskRun(record)
	assert.NoError(t, err)

	got, err := GetTaskRunByRunID(runId)
	assert.NoError(t, err)
	assert.Equal(t, runId, got.RunId)
	assert.Equal(t, "DONE", got.State)
	assert.Equal(t, 2, got.LoopCount)
	assert.Equal(t, "周报已生成", got.FinalResult)
	assert.Equal(t, int64(1700000000), got.StartTime)
	assert.Equal(t, int64(1700000100), got.EndTime)

	// steps/events JSON 可解码为前端可用的对象列表
	steps := TaskRunSteps(got.Steps)
	assert.Len(t, steps, 1)
	assert.Equal(t, "search", steps[0]["name"])
	assert.Equal(t, "done", steps[0]["status"])

	events := TaskRunEvents(got.Events)
	assert.Len(t, events, 1)
	assert.Equal(t, "run_started", events[0]["type"])

	// 损坏的 JSON 解码为空，不报错
	assert.Nil(t, TaskRunSteps("{broken"))
	assert.Nil(t, TaskRunEvents(""))
}

func TestTaskRunPageOrdering(t *testing.T) {
	// 同一 run_id 唯一，插入两条不同时间的记录验证倒序
	old := &TaskRunRecord{RunId: "page-old", State: "DONE", StartTime: 1700000000, EndTime: 1700000050}
	new := &TaskRunRecord{RunId: "page-new", State: "FAILED", StartTime: 1800000000, EndTime: 1800000050, ErrMsg: "boom"}
	assert.NoError(t, InsertTaskRun(old))
	assert.NoError(t, InsertTaskRun(new))

	runs, err := GetTaskRunsByPage(1, 50)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(runs), 2)

	// 最新开始时间在前
	foundNew, foundOld := false, false
	lastStart := int64(1 << 62)
	for _, run := range runs {
		assert.LessOrEqual(t, run.StartTime, lastStart)
		lastStart = run.StartTime
		if run.RunId == "page-new" {
			foundNew = true
		}
		if run.RunId == "page-old" {
			foundOld = true
		}
	}
	assert.True(t, foundNew)
	assert.True(t, foundOld)

	total, err := GetTaskRunsCount()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, 2)
}
