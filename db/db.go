package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "telegram_bot.db" // SQLite 数据库文件

var (
	DB *sql.DB
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Mode string `json:"mode"`
}

func InitTable() {
	var err error
	DB, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化表
	err = initializeTable(DB, "users")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("db initialize successfully")
}

// initializeTable check table exist or not.
func initializeTable(db *sql.DB, tableName string) error {
	// 查询表是否存在
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("table '%s' not exist，creating...\n", tableName)
			createTableSQL := `
				CREATE TABLE users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					mode VARCHAR(30) NOT NULL DEFAULT ''
				);`
			_, err := db.Exec(createTableSQL)
			if err != nil {
				return fmt.Errorf("create table fail: %v\n", err)
			}
			fmt.Print("create table success\n")
		} else {
			return fmt.Errorf("search table fail: %v\n", err)
		}
	} else {
		fmt.Printf("table '%s' exist\n", tableName)
	}

	return nil
}

// 插入新用户
func InsertUser(name, mode string) (int64, error) {
	// 插入数据
	insertSQL := `INSERT INTO users (name, mode) VALUES (?, ?)`
	result, err := DB.Exec(insertSQL, name, mode)
	if err != nil {
		return 0, err
	}

	// 获取插入的id
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// 根据 name 查询用户
func GetUserByName(name string) (*User, error) {
	// 查询单个用户，使用 name 作为条件
	querySQL := `SELECT id, name, mode FROM users WHERE name = ?`
	row := DB.QueryRow(querySQL, name)

	// 扫描查询结果并返回
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

// 读取所有用户
func GetUsers() ([]User, error) {
	// 查询所有用户
	rows, err := DB.Query("SELECT id, name, mode FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Mode); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// 检查是否有错误
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// 更新用户的模式
func UpdateUserMode(name string, mode string) error {
	// 更新用户模式
	updateSQL := `UPDATE users SET mode = ? WHERE name = ?`
	_, err := DB.Exec(updateSQL, mode, name)
	return err
}
