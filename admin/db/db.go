package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yincongcyincong/MuseBot/admin/conf"
	"github.com/yincongcyincong/MuseBot/logger"
)

const (
	sqlite3CreateTableSQL = `
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username VARCHAR(255) NOT NULL DEFAULT '',
				password VARCHAR(100) NOT NULL DEFAULT '',
				create_time int(10) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0'
			);
			CREATE TABLE bot (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				address VARCHAR(255) NOT NULL DEFAULT '',
				name VARCHAR(255) NOT NULL DEFAULT '',
				key_file TEXT NOT NULL,
				crt_file TEXT NOT NULL,
				ca_file TEXT NOT NULL,
				create_time int(10) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0',
				is_deleted int(10) NOT NULL DEFAULT '0'
			);
			insert into users values(1, 'admin', '21232f297a57a5a743894a0e4a801fc3', strftime('%s','now'), strftime('%s','now'))`
	
	mysqlCreateUsersSQL = `
			CREATE TABLE IF NOT EXISTS users (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
				username VARCHAR(255) NOT NULL DEFAULT '',
				password VARCHAR(100) NOT NULL DEFAULT '',
				create_time int(10) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0'
			);`
	
	mysqlCreateBotSQL = `
			CREATE TABLE IF NOT EXISTS bot (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
				address VARCHAR(255) NOT NULL DEFAULT '',
			    name VARCHAR(255) NOT NULL DEFAULT '',
				crt_file TEXT NOT NULL,
			    secret_file TEXT NOT NULL,
				create_time int(10) NOT NULL DEFAULT '0',
				update_time int(10) NOT NULL DEFAULT '0',
				is_deleted int(10) NOT NULL DEFAULT '0'
			);`
	MysqlInsertAdmin = `insert into users values(1, 'admin', '21232f297a57a5a743894a0e4a801fc3', Now(), Now())`
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
		
		if err := initializeMysqlTable(DB, "bot", mysqlCreateBotSQL); err != nil {
			logger.Fatal("create mysql table fail", "err", err)
		}
		
	}
	
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
		
		if tableName == "users" {
			_, err = db.Exec(MysqlInsertAdmin)
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
		if errors.Is(err, sql.ErrNoRows) {
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
