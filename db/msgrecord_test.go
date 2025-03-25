package db

import (
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"os"
	"strconv"
	"sync"
	"testing"
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
	userId := int64(1)
	MsgRecord = sync.Map{} // 清理数据

	aq := &AQ{Question: "What is Go?", Answer: "A programming language."}
	InsertMsgRecord(userId, aq, false)

	record := GetMsgRecord(userId)
	assert.NotNil(t, record, "Record should not be nil")
	assert.Equal(t, 1, len(record.AQs), "Should have 1 AQ pair")
	assert.Equal(t, "What is Go?", record.AQs[0].Question)
}

func TestInsertMsgRecord_ExceedLimit(t *testing.T) {
	userId := int64(1)
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
	userId := int64(1)
	MsgRecord = sync.Map{} // 清理数据

	aq := &AQ{Question: "Test Q", Answer: "Test A"}
	InsertMsgRecord(userId, aq, false)
	DeleteMsgRecord(userId)

	record := GetMsgRecord(userId)
	assert.Nil(t, record, "Record should be deleted")
}
