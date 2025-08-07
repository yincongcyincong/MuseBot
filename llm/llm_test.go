package llm

import (
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestSendMsg_WithMessageChan(t *testing.T) {
	msgChan := make(chan *param.MsgInfo, 1)
	l := &LLM{MessageChan: msgChan}
	msg := &param.MsgInfo{SendLen: 10}
	
	updated := l.SendMsg(msg, "hello")
	assert.Equal(t, "hello", updated.Content)
}

func TestSendMsg_WithHTTPMsgChan(t *testing.T) {
	httpChan := make(chan string, 1)
	l := &LLM{HTTPMsgChan: httpChan}
	
	l.SendMsg(&param.MsgInfo{}, "streamed text")
	select {
	case msg := <-httpChan:
		assert.Equal(t, "streamed text", msg)
	case <-time.After(time.Second):
		t.Error("Expected message on HTTPMsgChan")
	}
}

func TestOverLoop(t *testing.T) {
	l := &LLM{LoopNum: 9}
	assert.False(t, l.OverLoop())
	assert.Equal(t, 10, l.LoopNum)
	assert.True(t, l.OverLoop())
}

func TestNewLLM_DefaultsToClient(t *testing.T) {
	// This test assumes conf.BaseConfInfo.Type is properly mocked in actual tests
	// or indirectly validated via integration testing with each client type.
	l := NewLLM(
		WithUserId("u1"),
		WithContent("ask"),
		WithModel("m1"),
	)
	assert.NotNil(t, l)
	assert.Equal(t, "u1", l.UserId)
	assert.Equal(t, "ask", l.Content)
	assert.Equal(t, "m1", l.Model)
}
