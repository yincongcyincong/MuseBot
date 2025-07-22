// llm/llm_test.go
package llm_test

import (
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

func TestSendMsg_WithMessageChan(t *testing.T) {
	msgChan := make(chan *param.MsgInfo, 1)
	l := &llm.LLM{MessageChan: msgChan}
	msg := &param.MsgInfo{SendLen: 10}
	
	updated := l.SendMsg(msg, "hello")
	assert.Equal(t, "hello", updated.Content)
}

func TestSendMsg_WithHTTPMsgChan(t *testing.T) {
	httpChan := make(chan string, 1)
	l := &llm.LLM{HTTPMsgChan: httpChan}
	
	l.SendMsg(&param.MsgInfo{}, "streamed text")
	select {
	case msg := <-httpChan:
		assert.Equal(t, "streamed text", msg)
	case <-time.After(time.Second):
		t.Error("Expected message on HTTPMsgChan")
	}
}

func TestOverLoop(t *testing.T) {
	l := &llm.LLM{LoopNum: 4}
	assert.False(t, l.OverLoop())
	assert.Equal(t, 5, l.LoopNum)
	assert.True(t, l.OverLoop())
}

func TestNewLLM_DefaultsToClient(t *testing.T) {
	// This test assumes conf.BaseConfInfo.Type is properly mocked in actual tests
	// or indirectly validated via integration testing with each client type.
	l := llm.NewLLM(
		llm.WithUserId("u1"),
		llm.WithContent("ask"),
		llm.WithModel("m1"),
	)
	assert.NotNil(t, l)
	assert.Equal(t, "u1", l.UserId)
	assert.Equal(t, "ask", l.Content)
	assert.Equal(t, "m1", l.Model)
}
