package db

import (
	"database/sql"
	"testing"
	
	_ "github.com/mattn/go-sqlite3"
)

func TestInitializeSqlite3Table(t *testing.T) {
	// 使用 SQLite 内存数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory DB: %v", err)
	}
	defer db.Close()
	
	// 执行初始化
	err = initializeSqlite3Table(db)
	if err != nil {
		t.Errorf("initializeSqlite3Table failed: %v", err)
	}
	
	// 验证 users 表是否存在
	var name string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&name)
	if err != nil {
		t.Fatalf("Table 'users' not found: %v", err)
	}
	if name != "users" {
		t.Errorf("Expected table name 'users', got '%s'", name)
	}
}
