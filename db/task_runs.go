package db

import (
	"encoding/json"
	"fmt"

	"github.com/yincongcyincong/MuseBot/conf"
)

// TaskRunRecord 对应数据库中的 task_runs 表：一次 /task 执行的终态快照，
// 由 taskflow 包在运行结束时一次性写入；运行中的进度只存内存 Registry。
type TaskRunRecord struct {
	ID          int64  `json:"id"`
	RunId       string `json:"run_id"`
	UserId      string `json:"user_id"`
	ChatId      string `json:"chat_id"`
	Content     string `json:"content"`
	State       string `json:"state"`
	LoopCount   int    `json:"loop_count"`
	Steps       string `json:"steps"`  // JSON 数组字符串
	Events      string `json:"events"` // JSON 数组字符串
	FinalResult string `json:"final_result"`
	ErrMsg      string `json:"err_msg"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	FromBot     string `json:"from_bot"`
}

// InsertTaskRun 写入一次任务运行的终态快照
func InsertTaskRun(record *TaskRunRecord) error {
	insertSQL := `INSERT INTO task_runs (run_id, user_id, chat_id, content, state, loop_count, steps, events,
	                  final_result, err_msg, start_time, end_time, from_bot)
	              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := DB.Exec(insertSQL,
		record.RunId,
		record.UserId,
		record.ChatId,
		record.Content,
		record.State,
		record.LoopCount,
		record.Steps,
		record.Events,
		record.FinalResult,
		record.ErrMsg,
		record.StartTime,
		record.EndTime,
		conf.BaseConfInfo.BotName,
	)
	if err != nil {
		return fmt.Errorf("insert task run error: %w", err)
	}
	return nil
}

const taskRunSelectFields = "id, run_id, user_id, chat_id, content, state, loop_count, steps, events, final_result, err_msg, start_time, end_time, from_bot"

func scanTaskRun(scanner interface{ Scan(dest ...any) error }) (*TaskRunRecord, error) {
	var t TaskRunRecord
	if err := scanner.Scan(
		&t.ID, &t.RunId, &t.UserId, &t.ChatId, &t.Content, &t.State, &t.LoopCount,
		&t.Steps, &t.Events, &t.FinalResult, &t.ErrMsg, &t.StartTime, &t.EndTime, &t.FromBot,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTaskRunByRunID 按 run_id 查询历史快照
func GetTaskRunByRunID(runId string) (*TaskRunRecord, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM task_runs WHERE run_id = ? AND from_bot = ?", taskRunSelectFields)

	t, err := scanTaskRun(DB.QueryRow(querySQL, runId, conf.BaseConfInfo.BotName))
	if err != nil {
		return nil, fmt.Errorf("get task run by run_id error: %w", err)
	}
	return t, nil
}

// GetTaskRunsByPage 分页查询历史快照，按开始时间倒序
func GetTaskRunsByPage(page, pageSize int) ([]TaskRunRecord, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	listSQL := fmt.Sprintf(`
       SELECT %s
       FROM task_runs
       WHERE from_bot = ?
       ORDER BY start_time DESC, id DESC
       LIMIT ? OFFSET ?`, taskRunSelectFields)

	rows, err := DB.Query(listSQL, conf.BaseConfInfo.BotName, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query task runs by page error: %w", err)
	}
	defer rows.Close()

	taskRuns := make([]TaskRunRecord, 0)
	for rows.Next() {
		t, err := scanTaskRun(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task run row error: %w", err)
		}
		taskRuns = append(taskRuns, *t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return taskRuns, nil
}

// GetTaskRunsCount 历史快照总数
func GetTaskRunsCount() (int, error) {
	countSQL := "SELECT COUNT(*) FROM task_runs WHERE from_bot = ?"

	var count int
	if err := DB.QueryRow(countSQL, conf.BaseConfInfo.BotName).Scan(&count); err != nil {
		return 0, fmt.Errorf("get task runs count error: %w", err)
	}
	return count, nil
}

// TaskRunSteps 解码 steps JSON 字段；空/损坏时返回空列表
func TaskRunSteps(raw string) []map[string]interface{} {
	return decodeJSONSlice(raw)
}

// TaskRunEvents 解码 events JSON 字段；空/损坏时返回空列表
func TaskRunEvents(raw string) []map[string]interface{} {
	return decodeJSONSlice(raw)
}

func decodeJSONSlice(raw string) []map[string]interface{} {
	if raw == "" {
		return nil
	}
	list := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil
	}
	return list
}
