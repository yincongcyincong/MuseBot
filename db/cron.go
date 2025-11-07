package db

import (
	"fmt"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
)

// Cron 对应数据库中的 cron 表
type Cron struct {
	ID         int64  `json:"id"`
	CronName   string `json:"cron_name"`
	Type       string `json:"type"`
	CronSpec   string `json:"cron"` // 修正为 CronSpec，以避免与 Go 的关键字冲突
	TargetID   string `json:"target_id"`
	GroupID    string `json:"group_id"`
	Command    string `json:"command"`
	Prompt     string `json:"prompt"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	IsDeleted  int    `json:"is_deleted"`
	FromBot    string `json:"from_bot"`
	Status     int    `json:"status"`      // 0:disable 1:enable
	CronJobId  int64  `json:"cron_job_id"` // 定时任务在调度器中的 ID
}

// --- ➕ 增加 (Create/Insert) ---

// InsertCron 插入一条新的定时任务记录，默认状态为 1 (启用)，CronJobId 为 0
func InsertCron(cronName, cronSpec, targetID, groupID, command, prompt, t string) (int64, error) {
	insertSQL := `INSERT INTO cron (cron_name, cron, target_id, group_id, command, prompt, create_time, update_time, is_deleted, from_bot, status,
                  cron_job_id, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now().Unix()
	result, err := DB.Exec(insertSQL,
		cronName,
		cronSpec,
		targetID,
		groupID,
		command,
		prompt,
		now,
		now,
		0,
		*conf.BaseConfInfo.BotName,
		1, // 默认 status 为 1 (启用)
		0, // 默认 cron_job_id 为 0
		t,
	)
	if err != nil {
		return 0, fmt.Errorf("insert cron error: %w", err)
	}
	
	// 获取最后插入的ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id error: %w", err)
	}
	return id, nil
}

const cronSelectFields = "id, cron_name, type, cron, target_id, group_id, command, prompt, create_time, update_time, is_deleted, from_bot, status, cron_job_id"

func GetCronByID(id int64) (*Cron, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM cron WHERE id = ? and is_deleted = 0 and from_bot = ?", cronSelectFields)
	
	var c Cron
	err := DB.QueryRow(querySQL, id, *conf.BaseConfInfo.BotName).Scan(
		&c.ID, &c.CronName, &c.Type, &c.CronSpec, &c.TargetID, &c.GroupID, &c.Command, &c.Prompt, &c.CreateTime, &c.UpdateTime, &c.IsDeleted, &c.FromBot, &c.Status, &c.CronJobId,
	)
	
	if err != nil {
		return nil, fmt.Errorf("get cron by id error: %w", err)
	}
	return &c, nil
}

func GetActiveCrons() ([]*Cron, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM cron WHERE is_deleted = 0 and from_bot = ? ORDER BY id DESC", cronSelectFields)
	
	rows, err := DB.Query(querySQL, *conf.BaseConfInfo.BotName)
	
	if err != nil {
		return nil, fmt.Errorf("query active crons error: %w", err)
	}
	defer rows.Close()
	
	var crons []*Cron
	for rows.Next() {
		var c Cron
		if err := rows.Scan(
			&c.ID, &c.CronName, &c.Type, &c.CronSpec, &c.TargetID, &c.GroupID, &c.Command, &c.Prompt, &c.CreateTime, &c.UpdateTime, &c.IsDeleted, &c.FromBot, &c.Status, &c.CronJobId,
		); err != nil {
			return nil, fmt.Errorf("scan cron row error: %w", err)
		}
		crons = append(crons, &c)
	}
	
	// 检查 rows.Err()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return crons, nil
}

func GetCronsByPage(page, pageSize int, name string) ([]Cron, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	
	var (
		whereSQL = "WHERE is_deleted = 0 and from_bot = ?"
		args     = []interface{}{*conf.BaseConfInfo.BotName}
	)
	
	if name != "" {
		// 模糊匹配 cron_name
		whereSQL += " AND cron_name LIKE ?"
		args = append(args, "%"+name+"%")
	}
	
	// 查询数据，使用了 cronSelectFields
	listSQL := fmt.Sprintf(`
       SELECT %s
       FROM cron %s
       ORDER BY id DESC
       LIMIT ? OFFSET ?`, cronSelectFields, whereSQL)
	
	args = append(args, pageSize, offset)
	
	rows, err := DB.Query(listSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query crons by page error: %w", err)
	}
	defer rows.Close()
	
	var crons []Cron
	for rows.Next() {
		var c Cron
		if err := rows.Scan(
			&c.ID, &c.CronName, &c.Type, &c.CronSpec, &c.TargetID, &c.GroupID, &c.Command, &c.Prompt, &c.CreateTime, &c.UpdateTime, &c.IsDeleted, &c.FromBot, &c.Status, &c.CronJobId,
		); err != nil {
			return nil, fmt.Errorf("scan cron row error: %w", err)
		}
		crons = append(crons, c)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	
	return crons, nil
}

func GetCronsCount(name string) (int, error) {
	whereSQL := "WHERE is_deleted = 0 and from_bot = ?"
	args := []interface{}{*conf.BaseConfInfo.BotName}
	
	if name != "" {
		whereSQL += " AND cron_name LIKE ?"
		args = append(args, "%"+name+"%")
	}
	
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM cron %s", whereSQL)
	
	var count int
	err := DB.QueryRow(countSQL, args...).Scan(&count)
	
	if err != nil {
		return 0, fmt.Errorf("get crons count error: %w", err)
	}
	
	return count, nil
}

func UpdateCron(id int64, cronName, cronSpec, targetID, groupID, command, prompt, t string) error {
	updateSQL := `
        UPDATE cron
        SET cron_name = ?, cron = ?, target_id = ?, group_id = ?, command = ?, prompt = ?, update_time = ?
        WHERE id = ? AND is_deleted = 0 AND from_bot = ? AND type = ?`
	
	_, err := DB.Exec(updateSQL,
		cronName,
		cronSpec,
		targetID,
		groupID,
		command,
		prompt,
		time.Now().Unix(),
		id,
		*conf.BaseConfInfo.BotName,
		t,
	)
	return err
}

// UpdateCronStatus 更新定时任务的状态 (0:禁用, 1:启用)
func UpdateCronStatus(id int64, status int) error {
	updateSQL := `
        UPDATE cron
        SET status = ?, update_time = ?
        WHERE id = ? AND is_deleted = 0 AND from_bot = ?`
	
	_, err := DB.Exec(updateSQL,
		status,
		time.Now().Unix(),
		id,
		*conf.BaseConfInfo.BotName,
	)
	return err
}

// UpdateCronJobId 更新定时任务在调度器中的 Job ID
func UpdateCronJobId(id int64, cronJobID int) error {
	updateSQL := `
        UPDATE cron
        SET cron_job_id = ?, update_time = ?
        WHERE id = ? AND is_deleted = 0 AND from_bot = ?`
	
	_, err := DB.Exec(updateSQL,
		cronJobID,
		time.Now().Unix(),
		id,
		*conf.BaseConfInfo.BotName,
	)
	return err
}

// DeleteCronByID 对定时任务进行软删除（将 is_deleted 设为 1）
func DeleteCronByID(id int64) error {
	// 软删除：将 is_deleted 字段设置为 1
	deleteSQL := `UPDATE cron SET is_deleted = 1, update_time = ? WHERE id = ? AND from_bot = ?`
	
	_, err := DB.Exec(deleteSQL, time.Now().Unix(), id, *conf.BaseConfInfo.BotName)
	return err
}
