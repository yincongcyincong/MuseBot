package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
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
