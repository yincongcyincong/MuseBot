package http

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/taskflow"
	"github.com/yincongcyincong/MuseBot/utils"
)

// GetTaskList 返回任务执行列表：内存 Registry 中运行中的任务置顶，
// 之后接 task_runs 表里的历史快照（分页、按开始时间倒序）。
func GetTaskList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		logger.ErrorCtx(ctx, "parse form error", "err", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, err)
		return
	}
	page := utils.ParseInt(r.FormValue("page"))
	pageSize := utils.ParseInt(r.FormValue("page_size"))

	running := taskflow.RunningRuns()

	records, err := db.GetTaskRunsByPage(page, pageSize)
	if err != nil {
		logger.ErrorCtx(ctx, "get task runs error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	total, err := db.GetTaskRunsCount()
	if err != nil {
		logger.ErrorCtx(ctx, "get task runs count error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}

	// 历史里可能混有刚结束、Registry 尚未淘汰的 run，按 run_id 去重
	seen := make(map[string]bool, len(running))
	for _, run := range running {
		seen[run.RunId] = true
	}
	history := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		if seen[record.RunId] {
			continue
		}
		history = append(history, taskRunRecordView(&record))
	}

	utils.Success(ctx, w, r, map[string]interface{}{
		"running": running,
		"list":    history,
		"total":   total,
	})
}

// GetTaskRun 返回一次任务执行的完整快照（含步骤、事件与最终结果）：
// 优先取内存 Registry（运行中），不存在再查历史表。
func GetTaskRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	runId := r.URL.Query().Get("run_id")
	if runId == "" {
		utils.Failure(ctx, w, r, param.CodeParamError, "run_id is required", nil)
		return
	}

	if run := taskflow.GetRun(runId); run != nil {
		utils.Success(ctx, w, r, run)
		return
	}

	record, err := db.GetTaskRunByRunID(runId)
	if err != nil {
		// an unknown run_id (typo, or lost on restart) is not a server fault
		if errors.Is(err, sql.ErrNoRows) {
			utils.Failure(ctx, w, r, param.CodeParamError, "task run not found", nil)
			return
		}
		logger.ErrorCtx(ctx, "get task run error", "err", err)
		utils.Failure(ctx, w, r, param.CodeDBQueryFail, param.MsgDBQueryFail, err)
		return
	}
	utils.Success(ctx, w, r, taskRunRecordView(record))
}

// taskRunRecordView 把 DB 记录转成与内存快照同构的视图：steps/events
// 从 JSON 字符串解码成对象，前端拿到即可渲染。
func taskRunRecordView(record *db.TaskRunRecord) map[string]interface{} {
	return map[string]interface{}{
		"run_id":       record.RunId,
		"user_id":      record.UserId,
		"chat_id":      record.ChatId,
		"content":      record.Content,
		"state":        record.State,
		"loop":         record.LoopCount,
		"max_loops":    0,
		"plan":         db.TaskRunSteps(record.Steps),
		"events":       db.TaskRunEvents(record.Events),
		"final_result": record.FinalResult,
		"err_msg":      record.ErrMsg,
		"started_at":   record.StartTime,
		"finished_at":  record.EndTime,
		"from_db":      true,
	}
}
