package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	botUtils "github.com/yincongcyincong/MuseBot/utils"
)

const (
	sqlite3CreateTableSQL = `
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id varchar(100) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0',
				token int(20) NOT NULL DEFAULT '0',
				avail_token int(20) NOT NULL DEFAULT 0,
				create_time int(10) NOT NULL DEFAULT '0',
				from_bot VARCHAR(255) NOT NULL DEFAULT '',
				llm_config TEXT NOT NULL
			);
			CREATE TABLE records (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id varchar(100) NOT NULL DEFAULT '0',
				question TEXT NOT NULL,
				answer TEXT NOT NULL,
				content TEXT NOT NULL,
				create_time int(10) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0',
				is_deleted int(10) NOT NULL DEFAULT '0',
				token int(10) NOT NULL DEFAULT 0,
				mode VARCHAR(100) NOT NULL DEFAULT '',
				record_type tinyint(1) NOT NULL DEFAULT 0,
				from_bot VARCHAR(255) NOT NULL DEFAULT ''
			);
			CREATE TABLE rag_files (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				file_name VARCHAR(255) NOT NULL DEFAULT '',
				file_md5 VARCHAR(255) NOT NULL DEFAULT '',
				vector_id TEXT NOT NULL DEFAULT '',
				create_time int(10) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0',
				is_deleted int(10) NOT NULL DEFAULT '0',
				from_bot VARCHAR(255) NOT NULL DEFAULT ''
			);
			CREATE INDEX idx_users_user_id ON users(user_id);
			CREATE INDEX idx_records_user_id ON records(user_id);
			CREATE INDEX idx_records_create_time ON records(create_time);`
	
	mysqlCreateUsersSQL = `
			CREATE TABLE IF NOT EXISTS users (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
				user_id varchar(100) NOT NULL DEFAULT 0,
				update_time INT(10) NOT NULL DEFAULT 0,
				token BIGINT NOT NULL DEFAULT 0,
				avail_token BIGINT NOT NULL DEFAULT 0,
			    create_time int(10) NOT NULL DEFAULT '0',
			    from_bot VARCHAR(255) NOT NULL DEFAULT '',
			    llm_config TEXT NOT NULL
			);`
	
	mysqlCreateRecordsSQL = `
			CREATE TABLE IF NOT EXISTS records (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
				user_id varchar(100) NOT NULL DEFAULT 0,
				question MEDIUMTEXT NOT NULL,
				answer MEDIUMTEXT NOT NULL,
				content MEDIUMTEXT NOT NULL,
				create_time int(10) NOT NULL DEFAULT '0',
			    update_time int(10) NOT NULL DEFAULT '0',
				is_deleted int(10) NOT NULL DEFAULT '0',
				token int(10) NOT NULL DEFAULT 0,
			    mode VARCHAR(100) NOT NULL DEFAULT '',
			    record_type tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:text, 1:image 2:video 3: web',
			    from_bot VARCHAR(255) NOT NULL DEFAULT ''
			);`
	
	mysqlCreateRagFileSQL = `CREATE TABLE IF NOT EXISTS rag_files (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
				file_name VARCHAR(255) NOT NULL DEFAULT '',
				file_md5 VARCHAR(255) NOT NULL DEFAULT '',
    			vector_id TEXT NOT NULL,
				create_time int(10) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0',
				is_deleted int(10) NOT NULL DEFAULT '0',
    			from_bot VARCHAR(255) NOT NULL DEFAULT ''
			);`
	mysqlCreateUserIndexSQL = `CREATE INDEX idx_users_user_id ON users(user_id);`
	mysqlCreateIndexSQL     = `CREATE INDEX idx_records_user_id ON records(user_id);`
	mysqlCreateCTIndexSQL   = `CREATE INDEX idx_records_create_time ON records(create_time);`
)

var (
	DB *sql.DB
)

type DailyStat struct {
	Date     string `json:"date"`
	NewCount int    `json:"new_count"`
}

func InitTable() {
	var err error
	if _, err = os.Stat(botUtils.GetAbsPath("data")); os.IsNotExist(err) {
		// if dir don't exist, create it.
		err := os.MkdirAll(botUtils.GetAbsPath("data"), 0755)
		if err != nil {
			logger.Fatal("create direction fail:", "err", err)
			return
		}
		logger.Info("✅ create direction success")
	}
	
	DB, err = sql.Open(*conf.BaseConfInfo.DBType, *conf.BaseConfInfo.DBConf)
	if err != nil {
		logger.Fatal(err.Error())
	}
	
	// init table
	switch *conf.BaseConfInfo.DBType {
	case "sqlite3":
		err = initializeSqlite3Table(DB, "users")
		if err != nil {
			logger.Fatal("create sqlite table fail", "err", err)
		}
	case "mysql":
		// 检查并创建表
		if err := initializeMysqlTable(DB, "users", mysqlCreateUsersSQL); err != nil {
			logger.Fatal("create mysql table fail", "err", err)
		}
		
		if err := initializeMysqlTable(DB, "records", mysqlCreateRecordsSQL); err != nil {
			logger.Fatal("create mysql table fail", "err", err)
		}
		
		if err := initializeMysqlTable(DB, "rag_files", mysqlCreateRagFileSQL); err != nil {
			logger.Fatal("create mysql table fail", "err", err)
		}
	}
	
	InsertRecord(context.Background())
	
	logger.Info("db initialize successfully")
}

func initializeMysqlTable(db *sql.DB, tableName string, createSQL string) error {
	var tb string
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	err := db.QueryRow(query).Scan(&tb)
	
	if errors.Is(err, sql.ErrNoRows) || tb == "" {
		logger.Info("Table not exist, creating...", "tableName", tableName)
		_, err := db.Exec(createSQL)
		if err != nil {
			return fmt.Errorf("create table failed: %v", err)
		}
		logger.Info("Create table success", "tableName", tableName)
		
		// 创建索引（防止重复创建）
		if tableName == "records" {
			_, err = db.Exec(mysqlCreateUserIndexSQL)
			if err != nil {
				logger.Fatal("Create index failed", "err", err)
			}
			_, err = db.Exec(mysqlCreateIndexSQL)
			if err != nil {
				logger.Fatal("Create index failed", "err", err)
			}
			_, err = db.Exec(mysqlCreateCTIndexSQL)
			if err != nil {
				logger.Fatal("Create index failed", "err", err)
			}
		}
	} else if err != nil {
		return fmt.Errorf("search table failed: %v", err)
	} else {
		logger.Info("Table exists", "tableName", tableName)
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
			logger.Info("table not exist，creating...", "tableName", tableName)
			_, err := db.Exec(sqlite3CreateTableSQL)
			if err != nil {
				return fmt.Errorf("create table fail: %v", err)
			}
			logger.Info("create table success")
		} else {
			return fmt.Errorf("search table fail: %v", err)
		}
	} else {
		logger.Info("table exist", "tableName", tableName)
	}
	
	return nil
}
