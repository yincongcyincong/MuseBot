package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	botUtils "github.com/yincongcyincong/MuseBot/utils"
)

var (
	sqlite3TableSQLs = map[string]string{
		"users": `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id varchar(100) NOT NULL DEFAULT '0',
			update_time INTEGER NOT NULL DEFAULT '0',
			token INTEGER NOT NULL DEFAULT '0',
			avail_token INTEGER NOT NULL DEFAULT 0,
			create_time INTEGER NOT NULL DEFAULT '0',
			from_bot VARCHAR(255) NOT NULL DEFAULT '',
			llm_config TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id);
	`,
		"records": `
		CREATE TABLE IF NOT EXISTS records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id varchar(100) NOT NULL DEFAULT '0',
			question TEXT NOT NULL,
			answer TEXT NOT NULL,
			content TEXT NOT NULL,
			create_time INTEGER NOT NULL DEFAULT '0',
			update_time INTEGER NOT NULL DEFAULT '0',
			is_deleted INTEGER NOT NULL DEFAULT '0',
			token INTEGER NOT NULL DEFAULT 0,
			mode VARCHAR(100) NOT NULL DEFAULT '',
			record_type INTEGER NOT NULL DEFAULT 0, -- SQLite中用INTEGER代替tinyint
			from_bot VARCHAR(255) NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_records_user_id ON records(user_id);
		CREATE INDEX IF NOT EXISTS idx_records_create_time ON records(create_time);
	`,
		"rag_files": `
		CREATE TABLE IF NOT EXISTS rag_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name VARCHAR(255) NOT NULL DEFAULT '',
			file_md5 VARCHAR(255) NOT NULL DEFAULT '',
			vector_id TEXT NOT NULL DEFAULT '',
			create_time INTEGER NOT NULL DEFAULT '0',
			update_time INTEGER NOT NULL DEFAULT '0',
			is_deleted INTEGER NOT NULL DEFAULT '0',
			from_bot VARCHAR(255) NOT NULL DEFAULT ''
		);
	`,
		"cron": `
		CREATE TABLE IF NOT EXISTS cron (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cron_name VARCHAR(255) NOT NULL DEFAULT '',
			cron VARCHAR(255) NOT NULL DEFAULT '',
			target_id TEXT NOT NULL,
			group_id TEXT NOT NULL,
			command VARCHAR(255) NOT NULL DEFAULT '',
			prompt TEXT NOT NULL,
			status INTEGER NOT NULL DEFAULT 1, -- 0:disable 1:enable
			cron_job_id INTEGER NOT NULL DEFAULT '0',
			create_time INTEGER NOT NULL DEFAULT '0',
			update_time INTEGER NOT NULL DEFAULT '0',
			is_deleted INTEGER NOT NULL DEFAULT '0',
			from_bot VARCHAR(255) NOT NULL DEFAULT '',
		    type VARCHAR(255) NOT NULL DEFAULT '',
		    create_by VARCHAR(255) NOT NULL DEFAULT ''
		);
	`,
	}
	
	mysqlInitializeSQLs = []string{
		// 1. users 表 (嵌入索引)
		`
       CREATE TABLE IF NOT EXISTS users (
          id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
          user_id varchar(100) NOT NULL DEFAULT '0',
          update_time INT(10) NOT NULL DEFAULT 0,
          token BIGINT NOT NULL DEFAULT 0,
          avail_token BIGINT NOT NULL DEFAULT 0,
           create_time INT(10) NOT NULL DEFAULT 0,
           from_bot VARCHAR(255) NOT NULL DEFAULT '',
           llm_config TEXT NOT NULL,
           
           -- 嵌入索引：idx_users_user_id
           INDEX idx_users_user_id (user_id)
       ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`,
		// 2. records 表 (嵌入索引)
		`
       CREATE TABLE IF NOT EXISTS records (
          id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
          user_id varchar(100) NOT NULL DEFAULT '0',
          question MEDIUMTEXT NOT NULL,
          answer MEDIUMTEXT NOT NULL,
          content MEDIUMTEXT NOT NULL,
          create_time INT(10) NOT NULL DEFAULT 0,
           update_time INT(10) NOT NULL DEFAULT 0,
          is_deleted INT(10) NOT NULL DEFAULT 0,
          token INT(10) NOT NULL DEFAULT 0,
           mode VARCHAR(100) NOT NULL DEFAULT '',
           record_type tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:text, 1:image 2:video 3: web',
           from_bot VARCHAR(255) NOT NULL DEFAULT '',
           
           -- 嵌入索引：idx_records_user_id 和 idx_records_create_time
           INDEX idx_records_user_id (user_id),
           INDEX idx_records_create_time (create_time)
       ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`,
		// 3. rag_files 表 (无额外索引，仅PRIMARY KEY)
		`CREATE TABLE IF NOT EXISTS rag_files (
          id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
          file_name VARCHAR(255) NOT NULL DEFAULT '',
          file_md5 VARCHAR(255) NOT NULL DEFAULT '',
          vector_id TEXT NOT NULL,
          create_time INT(10) NOT NULL DEFAULT 0,
          update_time INT(10) NOT NULL DEFAULT 0,
          is_deleted INT(10) NOT NULL DEFAULT 0,
          from_bot VARCHAR(255) NOT NULL DEFAULT ''
       ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`,
		// 4. cron 表 (无额外索引，仅PRIMARY KEY)
		`CREATE TABLE IF NOT EXISTS cron (
          id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
          cron_name VARCHAR(255) NOT NULL DEFAULT '',
          cron VARCHAR(255) NOT NULL DEFAULT '',
          target_id TEXT NOT NULL,
          group_id TEXT NOT NULL,
          command VARCHAR(255) NOT NULL DEFAULT '',
          prompt TEXT NOT NULL,
          status tinyint(1) NOT NULL DEFAULT 1 COMMENT '0:disable 1:enable',
          cron_job_id INT(10) NOT NULL DEFAULT 0,
          create_time INT(10) NOT NULL DEFAULT 0,
          update_time INT(10) NOT NULL DEFAULT 0,
          is_deleted INT(10) NOT NULL DEFAULT 0,
          from_bot VARCHAR(255) NOT NULL DEFAULT '',
          type VARCHAR(255) NOT NULL DEFAULT '',
    	  create_by VARCHAR(255) NOT NULL DEFAULT ''
       ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`,
	}
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
		err = initializeSqlite3Table(DB)
		if err != nil {
			logger.Fatal("create sqlite table fail", "err", err)
		}
	case "mysql":
		err = initializeMySQLTables(DB)
		if err != nil {
			logger.Fatal("create mysql table fail", "err", err)
		}
	}
	
	InsertRecord(context.Background())
	
	logger.Info("db initialize successfully")
}

func initializeMySQLTables(db *sql.DB) error {
	for i, sqlStr := range mysqlInitializeSQLs {
		_, err := db.Exec(sqlStr)
		if err != nil {
			logger.Error("check table fail", "err", err)
			return fmt.Errorf("execute SQL batch %d fail: %v\nSQL: %s", i+1, err, sqlStr)
		}
	}
	
	return nil
}

// initializeSqlite3Table check table exist or not.
func initializeSqlite3Table(db *sql.DB) error {
	for tableName, createSQL := range sqlite3TableSQLs {
		_, err := db.Exec(createSQL)
		if err != nil {
			logger.Error("check table fail", "tableName", tableName, "err", err)
			return fmt.Errorf("create table %s fail: %v", tableName, err)
		}
	}
	
	return nil
}
