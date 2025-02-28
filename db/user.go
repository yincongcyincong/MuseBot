package db

import (
	"database/sql"
	"time"
)

// InsertUser insert user data
func InsertUser(name, mode string) (int64, error) {
	// insert data
	insertSQL := `INSERT INTO users (name, mode, updatetime) VALUES (?, ?, ?)`
	result, err := DB.Exec(insertSQL, name, mode, time.Now().Unix())
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

// 根据 name 查询用户
func GetUserByName(name string) (*User, error) {
	// select one use base on name
	querySQL := `SELECT id, name, mode FROM users WHERE name = ?`
	row := DB.QueryRow(querySQL, name)

	// scan row get result
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Mode)
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
	rows, err := DB.Query("SELECT id, name, mode, updatetime FROM users order by updatetime limit 1000")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Mode, &user.Updatetime); err != nil {
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
func UpdateUserMode(name string, mode string) error {
	updateSQL := `UPDATE users SET mode = ? WHERE name = ?`
	_, err := DB.Exec(updateSQL, mode, name)
	return err
}

// UpdateUserUpdateTime update user updateTime
func UpdateUserUpdateTime(name string, updateTime int64) error {
	updateSQL := `UPDATE users SET updateTime = ? WHERE name = ?`
	_, err := DB.Exec(updateSQL, updateTime, name)
	return err
}
