package db

import (
	"database/sql"
	"os"
	"strconv"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/admin/conf"
)

func TestMain(m *testing.M) {
	setup()
	
	code := m.Run()
	
	os.Exit(code)
}

func setup() {
	conf.InitConfig()
	InitTable()
}

func TestInitializeSqlite3Table(t *testing.T) {
	// 使用 SQLite 内存数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite memory DB: %v", err)
	}
	defer db.Close()
	
	// 执行初始化
	err = initializeSqlite3Table(db, "admin_users")
	if err != nil {
		t.Errorf("initializeSqlite3Table failed: %v", err)
	}
	
	// 验证 users 表是否存在
	var name string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='admin_users';`).Scan(&name)
	if err != nil {
		t.Fatalf("Table 'admin_users' not found: %v", err)
	}
	if name != "admin_users" {
		t.Errorf("Expected table name 'admin_users', got '%s'", name)
	}
}

func TestCreateAndGetBot(t *testing.T) {
	
	err := CreateBot("127.0.0.1", "BotA", "crt.pem", "key.pem", "ca.pem", "run")
	assert.NoError(t, err)
	
	// 检查数据库是否插入成功
	bot, err := GetBotByID("1")
	assert.NoError(t, err)
	assert.Equal(t, "BotA", bot.Name)
	assert.Equal(t, "127.0.0.1", bot.Address)
	assert.Equal(t, "run", bot.Command)
	
	DeleteAllBotData()
}

func TestUpdateBotAddress(t *testing.T) {
	
	err := CreateBot("127.0.0.1", "BotA", "crt.pem", "key.pem", "ca.pem", "run")
	assert.NoError(t, err)
	
	err = UpdateBotAddress(2, "192.168.0.1", "BotB", "crt2.pem", "key2.pem", "ca2.pem", "restart")
	assert.NoError(t, err)
	
	// 验证
	bot, err := GetBotByID("2")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.1", bot.Address)
	assert.Equal(t, "BotB", bot.Name)
	assert.Equal(t, "restart", bot.Command)
	
	DeleteAllBotData()
}

func TestSoftDeleteBot(t *testing.T) {
	
	err := CreateBot("127.0.0.1", "BotA", "crt.pem", "key.pem", "ca.pem", "run")
	assert.NoError(t, err)
	
	err = SoftDeleteBot(1)
	assert.NoError(t, err)
	
	// 再查时应找不到
	_, err = GetBotByID("1")
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	
	DeleteAllBotData()
}

func TestListBots_NoAddress(t *testing.T) {
	for i := 0; i < 3; i++ {
		err := CreateBot("10.0.0.1"+strconv.Itoa(i), "BotA"+strconv.Itoa(i), "crt.pem", "key.pem", "ca.pem", "run")
		assert.NoError(t, err)
	}
	
	bots, total, err := ListBots(0, 10, "")
	assert.NoError(t, err)
	
	assert.Equal(t, 3, len(bots))
	assert.Equal(t, 3, total)
	
	DeleteAllBotData()
}

func TestListBots_WithAddress(t *testing.T) {
	
	_ = CreateBot("127.0.0.1", "BotA", "crt.pem", "key.pem", "ca.pem", "run")
	_ = CreateBot("192.168.1.1", "BotB", "crt.pem", "key.pem", "ca.pem", "run")
	
	bots, total, err := ListBots(0, 10, "127")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(bots))
	assert.Equal(t, 1, total)
	assert.Equal(t, "127.0.0.1", bots[0].Address)
	
	DeleteAllBotData()
}
