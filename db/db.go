package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFile         = "./data/telegram_bot.db" // SQLite 数据库文件
	createTableSQL = `
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				mode VARCHAR(30) NOT NULL DEFAULT '',
				updatetime int(10) NOT NULL DEFAULT '0'
			);
			CREATE TABLE records (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				question TEXT NOT NULL,
				answer TEXT NOT NULL
			);
			CREATE INDEX idx_records_name ON users(name);`
)

var (
	DB *sql.DB
)

type User struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Mode       string `json:"mode"`
	Updatetime int64  `json:"updatetime"`
}

func InitTable() {
	var err error
	if _, err = os.Stat("./data"); os.IsNotExist(err) {
		// 文件夹不存在，创建它
		err := os.MkdirAll("./data", 0755)
		if err != nil {
			log.Fatal("❌ 创建文件夹失败:", err)
			return
		}
		fmt.Println("✅ 文件夹创建成功")
	}

	DB, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	// init table
	err = initializeTable(DB, "users")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("db initialize successfully")
}

// initializeTable check table exist or not.
func initializeTable(db *sql.DB, tableName string) error {
	// check table exist or not
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("table '%s' not exist，creating...\n", tableName)
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
