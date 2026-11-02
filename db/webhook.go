package db

import (
	"fmt"
	"time"

	"github.com/yincongcyincong/MuseBot/conf"
)

// Webhook 对应数据库中的 webhooks 表，事件触发的自动化任务；
// TokenHash 是触发 URL 中 token 的 SHA-256，明文 token 不落库
type Webhook struct {
	ID             int64  `json:"id"`
	WebhookName    string `json:"webhook_name"`
	TokenHash      string `json:"-"` // 凭据哈希，永远不序列化返回
	Source         string `json:"source"`
	Type           string `json:"type"`
	TargetID       string `json:"target_id"`
	GroupID        string `json:"group_id"`
	Command        string `json:"command"`
	PromptTemplate string `json:"prompt_template"`
	CreateTime     int64  `json:"create_time"`
	UpdateTime     int64  `json:"update_time"`
	IsDeleted      int    `json:"is_deleted"`
	FromBot        string `json:"from_bot"`
	Status         int    `json:"status"` // 0:disable 1:enable
	CreateBy       string `json:"create_by"`
}

// InsertWebhook 插入一条 webhook 记录，默认状态为 1 (启用)
func InsertWebhook(webhookName, tokenHash, source, t, targetID, groupID, command, promptTemplate, createBy string) (int64, error) {
	insertSQL := `INSERT INTO webhooks (webhook_name, token_hash, source, type, target_id, group_id, command,
	                  prompt_template, create_time, update_time, is_deleted, from_bot, status, create_by)
	              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now().Unix()
	result, err := DB.Exec(insertSQL,
		webhookName,
		tokenHash,
		source,
		t,
		targetID,
		groupID,
		command,
		promptTemplate,
		now,
		now,
		0,
		conf.BaseConfInfo.BotName,
		1, // 默认 status 为 1 (启用)
		createBy,
	)
	if err != nil {
		return 0, fmt.Errorf("insert webhook error: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id error: %w", err)
	}
	return id, nil
}

const webhookSelectFields = "id, webhook_name, token_hash, source, type, target_id, group_id, command, prompt_template, create_time, update_time, is_deleted, from_bot, status, create_by"

func scanWebhook(scanner interface{ Scan(dest ...any) error }) (*Webhook, error) {
	var w Webhook
	if err := scanner.Scan(
		&w.ID, &w.WebhookName, &w.TokenHash, &w.Source, &w.Type, &w.TargetID, &w.GroupID,
		&w.Command, &w.PromptTemplate, &w.CreateTime, &w.UpdateTime, &w.IsDeleted,
		&w.FromBot, &w.Status, &w.CreateBy,
	); err != nil {
		return nil, err
	}
	return &w, nil
}

// GetWebhookByID 按 ID 查询 webhook
func GetWebhookByID(id int64) (*Webhook, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM webhooks WHERE id = ? AND is_deleted = 0 AND from_bot = ?", webhookSelectFields)

	w, err := scanWebhook(DB.QueryRow(querySQL, id, conf.BaseConfInfo.BotName))
	if err != nil {
		return nil, fmt.Errorf("get webhook by id error: %w", err)
	}
	return w, nil
}

// GetWebhookByTokenHash 按触发 URL token 的 SHA-256 查询，只匹配本 bot 的启用记录
func GetWebhookByTokenHash(tokenHash string) (*Webhook, error) {
	querySQL := fmt.Sprintf("SELECT %s FROM webhooks WHERE token_hash = ? AND is_deleted = 0 AND from_bot = ? AND status = 1", webhookSelectFields)

	w, err := scanWebhook(DB.QueryRow(querySQL, tokenHash, conf.BaseConfInfo.BotName))
	if err != nil {
		return nil, fmt.Errorf("get webhook by token hash error: %w", err)
	}
	return w, nil
}

func GetWebhooksByPage(page, pageSize int, name string) ([]Webhook, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	whereSQL := "WHERE is_deleted = 0 AND from_bot = ?"
	args := []interface{}{conf.BaseConfInfo.BotName}

	if name != "" {
		// 模糊匹配 webhook_name
		whereSQL += " AND webhook_name LIKE ?"
		args = append(args, "%"+name+"%")
	}

	listSQL := fmt.Sprintf(`
       SELECT %s
       FROM webhooks %s
       ORDER BY id DESC
       LIMIT ? OFFSET ?`, webhookSelectFields, whereSQL)

	args = append(args, pageSize, offset)

	rows, err := DB.Query(listSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query webhooks by page error: %w", err)
	}
	defer rows.Close()

	webhooks := make([]Webhook, 0)
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, fmt.Errorf("scan webhook row error: %w", err)
		}
		webhooks = append(webhooks, *w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return webhooks, nil
}

func GetWebhooksCount(name string) (int, error) {
	whereSQL := "WHERE is_deleted = 0 AND from_bot = ?"
	args := []interface{}{conf.BaseConfInfo.BotName}

	if name != "" {
		whereSQL += " AND webhook_name LIKE ?"
		args = append(args, "%"+name+"%")
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM webhooks %s", whereSQL)

	var count int
	err := DB.QueryRow(countSQL, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get webhooks count error: %w", err)
	}

	return count, nil
}

// UpdateWebhook 更新 webhook 配置；token 不可更新，泄露后删除重建即可换新 token
func UpdateWebhook(id int64, webhookName, source, t, targetID, groupID, command, promptTemplate string) error {
	updateSQL := `
        UPDATE webhooks
        SET webhook_name = ?, source = ?, type = ?, target_id = ?, group_id = ?, command = ?, prompt_template = ?, update_time = ?
        WHERE id = ? AND is_deleted = 0 AND from_bot = ?`

	_, err := DB.Exec(updateSQL,
		webhookName,
		source,
		t,
		targetID,
		groupID,
		command,
		promptTemplate,
		time.Now().Unix(),
		id,
		conf.BaseConfInfo.BotName,
	)
	return err
}

// UpdateWebhookStatus 更新 webhook 状态 (0:禁用, 1:启用)
func UpdateWebhookStatus(id int64, status int) error {
	updateSQL := `
        UPDATE webhooks
        SET status = ?, update_time = ?
        WHERE id = ? AND is_deleted = 0 AND from_bot = ?`

	_, err := DB.Exec(updateSQL,
		status,
		time.Now().Unix(),
		id,
		conf.BaseConfInfo.BotName,
	)
	return err
}

// DeleteWebhookByID 软删除 webhook
func DeleteWebhookByID(id int64) error {
	deleteSQL := `UPDATE webhooks SET is_deleted = 1, update_time = ? WHERE id = ? AND from_bot = ?`

	_, err := DB.Exec(deleteSQL, time.Now().Unix(), id, conf.BaseConfInfo.BotName)
	return err
}
