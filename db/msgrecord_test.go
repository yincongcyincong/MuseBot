package db

import (
	"os"
	"strconv"
	"sync"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
)

func TestMain(m *testing.M) {
	setup()
	
	code := m.Run()
	
	os.Exit(code)
}

func setup() {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_bot_token")
	os.Setenv("DEEPSEEK_TOKEN", "test_deepseek_token")
	conf.InitConf()
	InitTable()
}

func TestInsertMsgRecord(t *testing.T) {
	userId := "1"
	MsgRecord = sync.Map{} // 清理数据
	
	aq := &AQ{Question: "What is Go?", Answer: "A programming language."}
	InsertMsgRecord(userId, aq, false)
	
	record := GetMsgRecord(userId)
	assert.NotNil(t, record, "Record should not be nil")
	assert.Equal(t, 1, len(record.AQs), "Should have 1 AQ pair")
	assert.Equal(t, "What is Go?", record.AQs[0].Question)
}

func TestInsertMsgRecord_ExceedLimit(t *testing.T) {
	userId := "1"
	MsgRecord = sync.Map{}
	
	for i := 0; i < MaxQAPair+5; i++ {
		aq := &AQ{Question: "Q" + strconv.Itoa(i), Answer: "A" + strconv.Itoa(i)}
		InsertMsgRecord(userId, aq, false)
	}
	
	record := GetMsgRecord(userId)
	assert.NotNil(t, record, "Record should not be nil")
	assert.Equal(t, MaxQAPair, len(record.AQs), "Should keep max limit AQ pairs")
}

func TestDeleteMsgRecord(t *testing.T) {
	userId := "1"
	MsgRecord = sync.Map{} // 清理数据
	
	aq := &AQ{Question: "Test Q", Answer: "Test A"}
	InsertMsgRecord(userId, aq, false)
	DeleteMsgRecord(userId)
	
	record := GetMsgRecord(userId)
	assert.Nil(t, record, "Record should be deleted")
}

func TestInsertRecordInfoAndGetRecords(t *testing.T) {
	
	userId := "12345"
	InsertUser(userId, "default")
	
	record := &Record{
		UserId:    userId,
		Question:  "What is AI?",
		Answer:    "AI is Artificial Intelligence.",
		Content:   "extra",
		Token:     5,
		IsDeleted: 0,
	}
	InsertRecordInfo(record)
	
	records, err := getRecordsByUserId(userId)
	if err != nil {
		t.Fatalf("getRecordsByUserId failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Question != "What is AI?" {
		t.Errorf("unexpected question: %s", records[0].Question)
	}
	
	DeleteMsgRecord(userId)
}

func TestDeleteRecord(t *testing.T) {
	
	userId := "456"
	InsertUser(userId, "default")
	
	// 插入未删除记录
	record := &Record{
		UserId:    userId,
		Question:  "Delete me?",
		Answer:    "Yes",
		Content:   "data",
		Token:     1,
		IsDeleted: 0,
	}
	InsertRecordInfo(record)
	
	// 删除
	err := DeleteRecord(userId)
	if err != nil {
		t.Fatalf("DeleteRecord failed: %v", err)
	}
	
	// 确认记录被标记为删除
	rows, _ := DB.Query("SELECT is_deleted FROM records WHERE user_id = ?", userId)
	defer rows.Close()
	
	for rows.Next() {
		var isDeleted int
		rows.Scan(&isDeleted)
		if isDeleted != 1 {
			t.Errorf("record not marked as deleted")
		}
	}
}
