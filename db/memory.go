package db

import (
	"context"
	"time"
)

// Memory is one durable fact about a user (preference, ongoing project,
// constraint) distilled out of the chat history and injected into future
// conversations' system context.
type Memory struct {
	ID         int64  `json:"id"`
	UserId     string `json:"user_id"`
	Content    string `json:"content"`
	CreateTime int64  `json:"create_time"`
}

// GetMemories returns the user's active memories, oldest first.
func GetMemories(userId string) []Memory {
	if DB == nil {
		return nil
	}
	rows, err := DB.Query(
		`SELECT id, user_id, content, create_time FROM memories
		 WHERE user_id = ? and is_deleted = 0 ORDER BY id ASC`, userId)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.UserId, &m.Content, &m.CreateTime); err != nil {
			return memories
		}
		memories = append(memories, m)
	}
	return memories
}

// ReplaceMemories swaps the user's memory set in one go: old rows are
// soft-deleted, then the new list inserted. The consolidation LLM owns the
// merge logic, so here everything is just written as given.
func ReplaceMemories(ctx context.Context, userId string, contents []string) error {
	if DB == nil {
		return nil
	}
	now := time.Now().Unix()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx,
		`UPDATE memories SET is_deleted = 1 WHERE user_id = ? and is_deleted = 0`, userId); err != nil {
		return err
	}
	for _, content := range contents {
		if content == "" {
			continue
		}
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO memories (user_id, content, create_time, is_deleted) VALUES (?, ?, ?, 0)`,
			userId, content, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}
