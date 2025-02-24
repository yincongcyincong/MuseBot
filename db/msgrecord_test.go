package db

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// 这里是测试前的初始化逻辑
	setup()

	// 运行测试
	code := m.Run()

	// 退出测试
	os.Exit(code)
}

func setup() {
	// 在这里执行你需要的初始化逻辑，比如连接数据库、设置环境变量等
	InitTable()
}

func TestInsertMsgRecord(t *testing.T) {
	username := "test_user"
	MsgRecord = sync.Map{} // 清理数据

	aq := &AQ{Question: "What is Go?", Answer: "A programming language."}
	InsertMsgRecord(username, aq, false)

	record := GetMsgRecord(username)
	assert.NotNil(t, record, "Record should not be nil")
	assert.Equal(t, 1, len(record.AQs), "Should have 1 AQ pair")
	assert.Equal(t, "What is Go?", record.AQs[0].Question)
}

func TestInsertMsgRecord_ExceedLimit(t *testing.T) {
	username := "test_user2"
	MsgRecord = sync.Map{} // 清理数据

	for i := 0; i < MaxQAPair+5; i++ {
		aq := &AQ{Question: "Q" + strconv.Itoa(i), Answer: "A" + strconv.Itoa(i)}
		InsertMsgRecord(username, aq, false)
	}

	record := GetMsgRecord(username)
	assert.NotNil(t, record, "Record should not be nil")
	assert.Equal(t, MaxQAPair, len(record.AQs), "Should keep max limit AQ pairs")
}

func TestDeleteMsgRecord(t *testing.T) {
	username := "test_user3"
	MsgRecord = sync.Map{} // 清理数据

	aq := &AQ{Question: "Test Q", Answer: "Test A"}
	InsertMsgRecord(username, aq, false)
	DeleteMsgRecord(username)

	record := GetMsgRecord(username)
	assert.Nil(t, record, "Record should be deleted")
}

func TestStarCheckUserLen(t *testing.T) {
	MsgRecord = sync.Map{} // 清理数据

	for i := 0; i < MaxUserLength+5; i++ {
		MsgRecord.Store("user"+strconv.Itoa(i), &MsgRecordInfo{updateTime: time.Now().Unix()})
	}

	UpdateDBData()

	totalNum := 0
	MsgRecord.Range(func(_, _ interface{}) bool {
		totalNum++
		return true
	})
	assert.LessOrEqual(t, totalNum, MaxUserLength, "Should clean up extra users")
}
