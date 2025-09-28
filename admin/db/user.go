package db

import (
	"database/sql"
	"time"
	
	"github.com/yincongcyincong/MuseBot/utils"
)

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

func CreateUser(username, password string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`INSERT INTO admin_users (username, password, create_time, update_time) VALUES (?, ?, ?, ?)`,
		username, utils.MD5(password), now, now)
	return err
}

func GetUserByID(id int) (*User, error) {
	row := DB.QueryRow(`SELECT id, username, create_time, update_time FROM admin_users WHERE id = ?`, id)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.CreateTime, &u.UpdateTime)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByUsername(username string) (*User, error) {
	row := DB.QueryRow(`SELECT id, username, password, create_time, update_time FROM admin_users WHERE username = ?`, username)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.CreateTime, &u.UpdateTime)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateUserPassword(id int, newPassword string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`UPDATE admin_users SET password = ?, update_time = ? WHERE id = ?`, utils.MD5(newPassword), now, id)
	return err
}

func DeleteUser(id int) error {
	_, err := DB.Exec(`DELETE FROM admin_users WHERE id = ?`, id)
	return err
}

func ListUsers(offset, limit int, username string) ([]User, int, error) {
	var (
		rows  *sql.Rows
		err   error
		args  []interface{}
		query string
	)
	
	users := make([]User, 0)
	
	// 构建查询 SQL
	if username != "" {
		query = `SELECT id, username, password, create_time, update_time FROM admin_users WHERE username LIKE ? LIMIT ? OFFSET ?`
		args = append(args, "%"+username+"%", limit, offset)
	} else {
		query = `SELECT id, username, password, create_time, update_time FROM admin_users LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}
	
	rows, err = DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.CreateTime, &u.UpdateTime)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	
	// 统计总数
	var countQuery string
	var countArgs []interface{}
	
	if username != "" {
		countQuery = `SELECT COUNT(*) FROM admin_users WHERE username LIKE ?`
		countArgs = append(countArgs, "%"+username+"%")
	} else {
		countQuery = `SELECT COUNT(*) FROM admin_users`
	}
	
	var total int
	err = DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}
