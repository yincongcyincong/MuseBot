package db

import (
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
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
	_, err := DB.Exec(`INSERT INTO users (username, password, create_time, update_time) VALUES (?, ?, ?, ?)`,
		username, utils.MD5(password), now, now)
	return err
}

func GetUserByID(id int) (*User, error) {
	row := DB.QueryRow(`SELECT id, username, create_time, update_time FROM users WHERE id = ?`, id)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.CreateTime, &u.UpdateTime)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByUsername(username string) (*User, error) {
	row := DB.QueryRow(`SELECT id, username, password, create_time, update_time FROM users WHERE username = ?`, username)
	u := &User{}
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.CreateTime, &u.UpdateTime)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateUserPassword(id int, newPassword string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`UPDATE users SET password = ?, update_time = ? WHERE id = ?`, utils.MD5(newPassword), now, id)
	return err
}

func DeleteUser(id int) error {
	_, err := DB.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

func ListUsers(offset, limit int) ([]User, int, error) {
	rows, err := DB.Query(`SELECT id, username, password, create_time, update_time FROM users LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	users := make([]User, 0)
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.CreateTime, &u.UpdateTime)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	
	var total int
	err = DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}
