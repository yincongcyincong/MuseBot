package db

import (
	"database/sql"
	"fmt"
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
)

type User struct {
	ID         int64  `json:"id"`
	UserId     string `json:"user_id"`
	Mode       string `json:"mode"`
	Token      int    `json:"token"`
	Updatetime int64  `json:"updatetime"`
	AvailToken int    `json:"avail_token"`
}

// InsertUser insert user data
func InsertUser(userId string, mode string) (int64, error) {
	// insert data
	insertSQL := `INSERT INTO users (user_id, mode, updatetime, avail_token) VALUES (?, ?, ?, ?)`
	result, err := DB.Exec(insertSQL, userId, mode, time.Now().Unix(), *conf.BaseConfInfo.TokenPerUser)
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

// GetUserByID get user by userId
func GetUserByID(userId string) (*User, error) {
	// select one use base on name
	querySQL := `SELECT id, user_id, mode, token, avail_token, updatetime FROM users WHERE user_id = ?`
	row := DB.QueryRow(querySQL, userId)
	
	// scan row get result
	var user User
	err := row.Scan(&user.ID, &user.UserId, &user.Mode, &user.Token, &user.AvailToken, &user.Updatetime)
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
func UpdateUserMode(userId string, mode string) error {
	updateSQL := `UPDATE users SET mode = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, mode, userId)
	return err
}

// UpdateUserUpdateTime update user updateTime
func UpdateUserUpdateTime(userId string, updateTime int64) error {
	updateSQL := `UPDATE users SET updatetime = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, updateTime, userId)
	return err
}

// UpdateUserToken update user token
func UpdateUserToken(userId string, token int) error {
	updateSQL := `UPDATE users SET token = token + ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, token, userId)
	return err
}

// AddAvailToken add token
func AddAvailToken(userId string, token int) error {
	updateSQL := `UPDATE users SET avail_token = avail_token + ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, token, userId)
	return err
}

func AddToken(userId string, token int) error {
	updateSQL := `UPDATE users SET token = token + ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, token, userId)
	return err
}

func GetUserByPage(page, pageSize int, userId string) ([]User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	
	// 构建 SQL
	var (
		whereSQL string
		args     []interface{}
	)
	
	if userId != "" {
		whereSQL = "WHERE user_id = ?"
		args = append(args, userId)
	}
	
	// 查询数据
	listSQL := fmt.Sprintf(`
		SELECT id, user_id, mode, token, updatetime, avail_token
		FROM users %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, whereSQL)
	args = append(args, pageSize, offset)
	
	rows, err := DB.Query(listSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.UserId, &u.Mode, &u.Token, &u.Updatetime, &u.AvailToken); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	
	return users, nil
}

func GetUserCount(userId string) (int, error) {
	var whereSQL string
	args := make([]interface{}, 0)
	
	if userId != "" {
		whereSQL = "WHERE user_id = ?"
		args = append(args, userId)
	}
	
	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereSQL)
	var total int
	if err := DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return 0, err
	}
	
	return total, nil
}
