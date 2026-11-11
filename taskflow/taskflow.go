// Package taskflow 记录 /task 多智能体工作流的执行过程：llm 包在状态机
// 各节点与工具调用处上报结构化事件，HTTP/admin 层读取快照，前端渲染流程图。
// 运行中的任务只存内存 Registry（有上限），结束(终态)时一次性落库 task_runs。
package taskflow

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/utils"
)

// 任务/步骤状态
const (
	StatePlan     = "PLAN"
	StateExecute  = "EXECUTE"
	StateReflect  = "REFLECT"
	StateSummarize = "SUMMARIZE"
	StateDone     = "DONE"
	StateFailed   = "FAILED"
	StateCanceled = "CANCELED"

	StepPending = "pending"
	StepRunning = "running"
	StepDone    = "done"
	StepFailed  = "failed"
)

// TaskStep 是计划中的一个步骤（agent/skill）
type TaskStep struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Result      string `json:"result"`
}

// TaskEvent 是 append-only 时间线中的一条事件
type TaskEvent struct {
	Ts     int64  `json:"ts"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
}

// TaskRun 是一次 /task 执行的完整快照；任务 goroutine 写、HTTP 读，
// 用 mu 保护。JSON 序列化走 Snapshot 返回的副本。
type TaskRun struct {
	mu sync.RWMutex `json:"-"`

	RunId       string     `json:"run_id"`
	UserId      string     `json:"user_id"`
	ChatId      string     `json:"chat_id"`
	Content     string     `json:"content"`
	State       string     `json:"state"`
	Loop        int        `json:"loop"`
	MaxLoops    int        `json:"max_loops"`
	Plan        []TaskStep `json:"plan"`
	StartedAt   int64      `json:"started_at"`
	FinishedAt  int64      `json:"finished_at"`
	FinalResult string     `json:"final_result"`
	ErrMsg      string     `json:"err_msg"`
	Events      []TaskEvent `json:"events"`
}

func (t *TaskRun) terminal() bool {
	switch t.State {
	case StateDone, StateFailed, StateCanceled:
		return true
	}
	return false
}

// Snapshot 返回可安全跨 goroutine 使用的深副本（逐字段构造，避免拷贝锁）
func (t *TaskRun) Snapshot() *TaskRun {
	if t == nil {
		return nil
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	clone := &TaskRun{
		RunId:       t.RunId,
		UserId:      t.UserId,
		ChatId:      t.ChatId,
		Content:     t.Content,
		State:       t.State,
		Loop:        t.Loop,
		MaxLoops:    t.MaxLoops,
		StartedAt:   t.StartedAt,
		FinishedAt:  t.FinishedAt,
		FinalResult: t.FinalResult,
		ErrMsg:      t.ErrMsg,
	}
	clone.Plan = append([]TaskStep(nil), t.Plan...)
	clone.Events = append([]TaskEvent(nil), t.Events...)
	return clone
}

const (
	// maxEvents 限制单个 run 的事件条数，超出丢最旧的
	maxEvents = 200
	// maxRuns 限制内存 Registry 中保留的 run 总数，超出清最旧的终态 run
	maxRuns = 500
	// detailMaxLen 单条事件/结果文本的截断长度
	detailMaxLen = 400
)

func truncate(s string) string {
	return utils.TruncateUTF8(s, detailMaxLen)
}

// ---------------------------------------------------------------- Tracker

// Tracker 由 llm 包持有并上报进度。所有方法都是 nil-receiver 安全的：
// 未注入跟踪（非 /task 场景）时零开销空操作。
type Tracker struct {
	run *TaskRun
}

// StartRun 注册一次新的任务运行，返回注入 LLMTaskReq 的 Tracker
func StartRun(content, userId, chatId string) *Tracker {
	run := &TaskRun{
		RunId:     uuid.New().String(),
		UserId:    userId,
		ChatId:    chatId,
		Content:   truncate(content),
		State:     StatePlan,
		StartedAt: time.Now().Unix(),
	}
	registry.Store(run.RunId, run)
	pruneRuns()
	return &Tracker{run: run}
}

func (t *Tracker) RunStarted() {
	if t == nil || t.run == nil {
		return
	}
	t.append("run_started", t.run.Content)
}

// Stage 上报状态机阶段切换（PLAN/EXECUTE/REFLECT/SUMMARIZE）
func (t *Tracker) Stage(state string) {
	if t == nil || t.run == nil {
		return
	}
	t.run.mu.Lock()
	t.run.State = state
	t.run.mu.Unlock()
	t.append("stage", state)
}

// LoopStarted 上报进入第几轮 EXECUTE 循环（loop 从 1 开始）
func (t *Tracker) LoopStarted(loop, maxLoops int) {
	if t == nil || t.run == nil {
		return
	}
	t.run.mu.Lock()
	t.run.Loop = loop
	t.run.MaxLoops = maxLoops
	t.run.mu.Unlock()
	t.append("loop_started", fmt.Sprintf("loop %d/%d", loop, maxLoops))
}

// PlanCreated 上报一次完整计划（初始规划或 REFLECT 后重规划，整体替换）
func (t *Tracker) PlanCreated(steps []TaskStep) {
	if t == nil || t.run == nil {
		return
	}
	t.run.mu.Lock()
	t.run.Plan = steps
	t.run.mu.Unlock()
	t.append("plan_created", fmt.Sprintf("%d steps", len(steps)))
}

// StepStarted 上报第 idx 个步骤开始执行
func (t *Tracker) StepStarted(idx int, name string) {
	if t == nil || t.run == nil {
		return
	}
	t.run.mu.Lock()
	if idx >= 0 && idx < len(t.run.Plan) {
		t.run.Plan[idx].Status = StepRunning
	}
	t.run.mu.Unlock()
	t.append("step_started", fmt.Sprintf("#%d %s", idx+1, name))
}

// StepFinished 上报第 idx 个步骤结束（err != nil 标记失败）
func (t *Tracker) StepFinished(idx int, name, result string, err error) {
	if t == nil || t.run == nil {
		return
	}
	status := StepDone
	detail := fmt.Sprintf("#%d %s ok", idx+1, name)
	if err != nil {
		status = StepFailed
		detail = fmt.Sprintf("#%d %s failed: %v", idx+1, name, err)
	}
	t.run.mu.Lock()
	if idx >= 0 && idx < len(t.run.Plan) {
		t.run.Plan[idx].Status = status
		t.run.Plan[idx].Result = truncate(result)
	}
	t.run.mu.Unlock()
	t.append("step_finished", detail)
}

// ToolCalled 上报一次工具调用（内置工具与 MCP 统一入口处）
func (t *Tracker) ToolCalled(name, args string, err error, duration time.Duration) {
	if t == nil || t.run == nil {
		return
	}
	detail := fmt.Sprintf("%s (%s) args=%s", name, duration.Round(time.Millisecond), truncate(args))
	if err != nil {
		detail = fmt.Sprintf("%s (%s) failed: %v", name, duration.Round(time.Millisecond), err)
	}
	t.append("tool_called", detail)
}

// ApprovalRequested 上报一次需要用户授权的操作
func (t *Tracker) ApprovalRequested(kind, target string) {
	if t == nil || t.run == nil {
		return
	}
	t.append("approval_requested", fmt.Sprintf("%s %s", kind, target))
}

// Finish 上报终态：更新快照、落库、并按上限清理 Registry
func (t *Tracker) Finish(state, finalResult, errMsg string) {
	if t == nil || t.run == nil {
		return
	}
	t.run.mu.Lock()
	t.run.State = state
	t.run.FinishedAt = time.Now().Unix()
	t.run.FinalResult = finalResult
	t.run.ErrMsg = truncate(errMsg)
	run := t.run
	t.run.mu.Unlock()
	t.append("run_finished", state)

	persistRun(run)
}

// ---------------------------------------------------------------- Registry

var registry sync.Map // runId -> *TaskRun

// GetRun 按 runId 取运行中/近期完成的快照；不存在返回 nil
func GetRun(runId string) *TaskRun {
	value, ok := registry.Load(runId)
	if !ok {
		return nil
	}
	return value.(*TaskRun).Snapshot()
}

// RunningRuns 返回所有非终态 run（最新在前）
func RunningRuns() []*TaskRun {
	runs := make([]*TaskRun, 0)
	registry.Range(func(_, value any) bool {
		run := value.(*TaskRun)
		if !run.terminal() {
			runs = append(runs, run.Snapshot())
		}
		return true
	})
	// 按开始时间倒序
	for i := 0; i < len(runs); i++ {
		for j := i + 1; j < len(runs); j++ {
			if runs[j].StartedAt > runs[i].StartedAt {
				runs[i], runs[j] = runs[j], runs[i]
			}
		}
	}
	return runs
}

func (t *Tracker) append(eventType, detail string) {
	if t == nil || t.run == nil {
		return
	}
	t.run.mu.Lock()
	t.run.Events = append(t.run.Events, TaskEvent{
		Ts:     time.Now().Unix(),
		Type:   eventType,
		Detail: truncate(detail),
	})
	if len(t.run.Events) > maxEvents {
		t.run.Events = t.run.Events[len(t.run.Events)-maxEvents:]
	}
	t.run.mu.Unlock()
}

// persistRun 把终态快照写入 task_runs 表；失败只记日志不影响任务本身
func persistRun(run *TaskRun) {
	snapshot := run.Snapshot()

	steps, _ := json.Marshal(snapshot.Plan)
	events, _ := json.Marshal(snapshot.Events)

	err := db.InsertTaskRun(&db.TaskRunRecord{
		RunId:       snapshot.RunId,
		UserId:      snapshot.UserId,
		ChatId:      snapshot.ChatId,
		Content:     snapshot.Content,
		State:       snapshot.State,
		LoopCount:   snapshot.Loop,
		Steps:       string(steps),
		Events:      string(events),
		FinalResult: snapshot.FinalResult,
		ErrMsg:      snapshot.ErrMsg,
		StartTime:   snapshot.StartedAt,
		EndTime:     snapshot.FinishedAt,
	})
	if err != nil {
		logger.Warn("persist task run fail", "runId", snapshot.RunId, "err", err)
	}
}

// pruneRuns 超过 maxRuns 时优先清最旧的终态 run；若全部是运行中的
// （理论上不可能超过并发上限），淘汰最旧的
func pruneRuns() {
	size := 0
	registry.Range(func(_, _ any) bool {
		size++
		return true
	})
	if size <= maxRuns {
		return
	}

	type entry struct {
		id        string
		startedAt int64
		terminal  bool
	}
	entries := make([]entry, 0, size)
	registry.Range(func(key, value any) bool {
		run := value.(*TaskRun)
		run.mu.RLock()
		entries = append(entries, entry{id: key.(string), startedAt: run.StartedAt, terminal: run.terminal()})
		run.mu.RUnlock()
		return true
	})

	// 终态优先淘汰，其次按开始时间
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			rank := func(e entry) int {
				if e.terminal {
					return 0
				}
				return 1
			}
			if rank(entries[j]) < rank(entries[i]) ||
				(rank(entries[j]) == rank(entries[i]) && entries[j].startedAt < entries[i].startedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	for _, e := range entries {
		if size <= maxRuns {
			break
		}
		registry.Delete(e.id)
		size--
	}
}
