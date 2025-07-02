package db

import (
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/utils"
)

type User struct {
	ID         int
	Username   string
	Password   string
	CreateTime int64
	UpdateTime int64
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
