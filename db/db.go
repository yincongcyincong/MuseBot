package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
)

const (
	sqlite3CreateTableSQL = `
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id int(11) NOT NULL DEFAULT '0',
				mode VARCHAR(30) NOT NULL DEFAULT '',
				updatetime int(10) NOT NULL DEFAULT '0'
			);
			CREATE TABLE records (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id int(11) NOT NULL DEFAULT '0',
				question TEXT NOT NULL,
				answer TEXT NOT NULL
			);
			CREATE INDEX idx_records_user_id ON users(user_id);`
	mysqlCreateUsersSQL = `
CREATE TABLE IF NOT EXISTS users (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT(11) NOT NULL DEFAULT 0,
    mode VARCHAR(30) NOT NULL DEFAULT '',
    updatetime INT(10) NOT NULL DEFAULT 0
);`

	mysqlCreateRecordsSQL = `
CREATE TABLE IF NOT EXISTS records (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT(11) NOT NULL DEFAULT 0,
    question TEXT NOT NULL,
    answer TEXT NOT NULL
);`

	mysqlCreateIndexSQL = `CREATE INDEX idx_records_user_id ON records(user_id);`
)

var (
	DB *sql.DB
)

func InitTable() {
	var err error
	if _, err = os.Stat("./data"); os.IsNotExist(err) {
		// if dir don't exist, create it.
		err := os.MkdirAll("./data", 0755)
		if err != nil {
			log.Fatal("❌ 创建文件夹失败:", err)
			return
		}
		fmt.Println("✅ 文件夹创建成功")
	}

	DB, err = sql.Open(*conf.DBType, *conf.DBConf)
	if err != nil {
		log.Fatal(err)
	}

	// init table
	switch *conf.DBType {
	case "sqlite3":
		err = initializeSqlite3Table(DB, "users")
		if err != nil {
			log.Fatal(err)
		}
	case "mysql":
		// 检查并创建表
		if err := initializeMysqlTable(DB, "users", mysqlCreateUsersSQL); err != nil {
			log.Fatal(err)
		}

		if err := initializeMysqlTable(DB, "records", mysqlCreateRecordsSQL); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("db initialize successfully")
}

func initializeMysqlTable(db *sql.DB, tableName string, createSQL string) error {
	var tb string
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	err := db.QueryRow(query).Scan(&tb)

	// 如果表不存在，则创建
	if errors.Is(err, sql.ErrNoRows) || tb == "" {
		fmt.Printf("Table '%s' not exist, creating...\n", tableName)
		_, err := db.Exec(createSQL)
		if err != nil {
			return fmt.Errorf("Create table failed: %v", err)
		}
		fmt.Println("Create table success:", tableName)

		// 创建索引（防止重复创建）
		_, err = DB.Exec(mysqlCreateIndexSQL)
		if err != nil {
			log.Fatal("Create index failed:", err)
		} else {
			fmt.Println("Create index success")
		}
	} else if err != nil {
		return fmt.Errorf("Search table failed: %v", err)
	} else {
		fmt.Printf("Table '%s' exists\n", tableName)
	}

	return nil
}

// initializeSqlite3Table check table exist or not.
func initializeSqlite3Table(db *sql.DB, tableName string) error {
	// check table exist or not
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("table '%s' not exist，creating...\n", tableName)
			_, err := db.Exec(sqlite3CreateTableSQL)
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
