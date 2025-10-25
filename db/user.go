package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

type User struct {
	ID           int64            `json:"id"`
	UserId       string           `json:"user_id"`
	Token        int              `json:"token"`
	UpdateTime   int64            `json:"update_time"`
	CreateTime   int64            `json:"create_time"`
	AvailToken   int              `json:"avail_token"`
	LLMConfig    string           `json:"llm_config"`
	LLMConfigRaw *param.LLMConfig `json:"llm_config_raw"`
}

// InsertUser insert user data
func InsertUser(userId string, llmConfig string) (int64, error) {
	userInfo, err := GetUserByID(userId)
	if err != nil {
		return 0, err
	}
	if userInfo != nil && userInfo.ID != 0 {
		return userInfo.ID, nil
	}
	
	// insert data
	insertSQL := `INSERT INTO users (user_id, llm_config, update_time, create_time, avail_token, from_bot) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := DB.Exec(insertSQL, userId, llmConfig, time.Now().Unix(), time.Now().Unix(), *conf.BaseConfInfo.TokenPerUser, *conf.BaseConfInfo.BotName)
	if err != nil {
		return 0, err
	}
	
	// get last insert id
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUserByID get user by userId
func GetUserByID(userId string) (*User, error) {
	// select one use base on name
	querySQL := `SELECT id, user_id, llm_config, token, avail_token, update_time, create_time FROM users WHERE user_id = ?`
	row := DB.QueryRow(querySQL, userId)
	
	// scan row get result
	var user User
	err := row.Scan(&user.ID, &user.UserId, &user.LLMConfig, &user.Token, &user.AvailToken, &user.UpdateTime, &user.CreateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到数据，返回 nil
			return nil, nil
		}
		return nil, err
	}
	
	if user.LLMConfig != "" {
		err := json.Unmarshal([]byte(user.LLMConfig), &user.LLMConfigRaw)
		if err != nil {
			return nil, fmt.Errorf("UnmarshalJSON failed: %v", err)
		}
	}
	
	return &user, nil
}

// GetUsers get 1000 users order by updatetime
func GetUsers() ([]User, error) {
	rows, err := DB.Query("SELECT id, user_id, llm_config, update_time FROM users order by update_time limit 10000")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.UserId, &user.LLMConfig, &user.UpdateTime); err != nil {
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

// UpdateUserLLMConfig update user llm config
func UpdateUserLLMConfig(userId string, llmConfig string) error {
	updateSQL := `UPDATE users SET llm_config = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, llmConfig, userId)
	return err
}

// UpdateUserUpdateTime update user updateTime
func UpdateUserUpdateTime(userId string, updateTime int64) error {
	updateSQL := `UPDATE users SET update_time = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, updateTime, userId)
	return err
}

// AddAvailToken add token
func AddAvailToken(userId string, token int) error {
	updateSQL := `UPDATE users SET avail_token = avail_token + ?, update_time = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, token, time.Now().Unix(), userId)
	return err
}

func AddToken(userId string, token int) error {
	updateSQL := `UPDATE users SET token = token + ?, update_time = ? WHERE user_id = ?`
	_, err := DB.Exec(updateSQL, token, time.Now().Unix(), userId)
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
		SELECT id, user_id, llm_config, token, update_time, avail_token, create_time
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
		if err := rows.Scan(&u.ID, &u.UserId, &u.LLMConfig, &u.Token, &u.UpdateTime, &u.AvailToken, &u.CreateTime); err != nil {
			return nil, err
		}
		if u.LLMConfig != "" {
			err := json.Unmarshal([]byte(u.LLMConfig), &u.LLMConfigRaw)
			if err != nil {
				return nil, fmt.Errorf("UnmarshalJSON failed: %v", err)
			}
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

func GetDailyNewUsers(days int) ([]DailyStat, error) {
	var query string
	var intervalSeconds int64
	
	if days <= 3 {
		intervalSeconds = 3600 // 每小时
	} else if days <= 7 {
		intervalSeconds = 3 * 3600 // 每3小时
	} else {
		intervalSeconds = 86400 // 每天
	}
	
	if *conf.BaseConfInfo.DBType == "mysql" {
		query = `
			SELECT
				FLOOR(create_time / ?) * ? AS time_group,
				COUNT(DISTINCT user_id) AS new_count
			FROM users
			WHERE create_time >= UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL ? DAY))
			GROUP BY time_group
			ORDER BY time_group DESC;
		`
	} else if *conf.BaseConfInfo.DBType == "sqlite3" {
		query = `
			SELECT
				(create_time / ?) * ? AS time_group,
				COUNT(DISTINCT user_id) AS new_count
			FROM users
			WHERE create_time >= strftime('%s', date('now', ? || ' days'))
			GROUP BY time_group
			ORDER BY time_group DESC;
		`
	} else {
		return nil, fmt.Errorf("unsupported DBType: %s", *conf.BaseConfInfo.DBType)
	}
	
	var rows *sql.Rows
	var err error
	if *conf.BaseConfInfo.DBType == "sqlite3" {
		rows, err = DB.Query(query, intervalSeconds, intervalSeconds, -days)
	} else {
		rows, err = DB.Query(query, intervalSeconds, intervalSeconds, days)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stats []DailyStat
	for rows.Next() {
		var stat DailyStat
		if err := rows.Scan(&stat.Date, &stat.NewCount); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	
	return stats, nil
}

func GetCtxUserInfo(ctx context.Context) *User {
	userInfo, ok := ctx.Value("user_info").(*User)
	if ok {
		return userInfo
	}
	
	return nil
}
