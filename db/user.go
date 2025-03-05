package db

import (
	"database/sql"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"time"
)

type User struct {
	ID         int64  `json:"id"`
	UserId     int64  `json:"user_id"`
	Mode       string `json:"mode"`
	Updatetime int64  `json:"updatetime"`
}

// InsertUser insert user data
func InsertUser(userId int64, mode string) (int64, error) {
	// insert data
	insertSQL := `INSERT INTO users (user_id, mode, updatetime) VALUES (?, ?, ?)`
	result, err := DB.Exec(insertSQL, userId, mode, time.Now().Unix())
	if err != nil {
		return 0, err
	}

	// get last insert id
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	metrics.TotalUsers.Inc()
	return id, nil
}

// 根据 name 查询用户
func GetUserByID(userId int64) (*User, error) {
	// select one use base on name
	querySQL := `SELECT id, user_id, mode FROM users WHERE user_id = ?`
	row := DB.QueryRow(querySQL, userId)

	// scan row get result
	var user User
	err := row.Scan(&user.ID, &user.UserId, &user.Mode)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到数据，返回 nil
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUsers get 1000 users order by updatetime
func GetUsers() ([]User, error) {
	rows, err := DB.Query("SELECT id, user_id, mode, updatetime FROM users order by updatetime limit 1000")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.UserId, &user.Mode, &user.Updatetime); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// check error
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserMode update user mode
func UpdateUserMode(userId int64, mode string) error {
	updateSQL := `UPDATE users SET mode = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, mode, userId)
	return err
}

// UpdateUserUpdateTime update user updateTime
func UpdateUserUpdateTime(userId int64, updateTime int64) error {
	updateSQL := `UPDATE users SET updateTime = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, updateTime, userId)
	return err
}
