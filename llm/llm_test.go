package llm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/db"
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

// A task planner without any channel attached must not block forever on a nil
// channel; the content is kept in WholeContent instead.
func TestSendMsg_WithoutChanDoesNotBlock(t *testing.T) {
	l := &LLM{}

	done := make(chan struct{})
	go func() {
		l.SendMsg(&param.MsgInfo{}, "first")
		l.SendMsg(&param.MsgInfo{}, " second")
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, "first second", l.WholeContent)
	case <-time.After(time.Second):
		t.Error("SendMsg blocked on nil channel")
	}
}

func TestOverLoop(t *testing.T) {
	l := &LLM{LoopNum: MostLoop - 1}
	assert.False(t, l.OverLoop())
	assert.Equal(t, MostLoop, l.LoopNum)
	assert.True(t, l.OverLoop())
}

func TestNewLLM_DefaultsToClient(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	l := NewLLM(
		WithUserId("u1"),
		WithContent("ask"),
		WithModel("m1"),
		WithContext(ctx),
	)
	assert.NotNil(t, l)
	assert.Equal(t, "u1", l.UserId)
	assert.Equal(t, "ask", l.Content)
	assert.Equal(t, "m1", l.Model)
}
