package llm

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

func TestGetModel(t *testing.T) {
	l := &LLM{UserId: "test-user"}
	h := &VolReq{}
	h.GetModel(l)
	
	assert.Equal(t, param.ModelDeepSeekR1_528, l.Model)
}

func TestGetMessages_NoHistory(t *testing.T) {
	
	h := &VolReq{}
	h.GetMessages("no-history-user", "hello")
	assert.Len(t, h.VolMsgs, 1)
	assert.Equal(t, "hello", *h.VolMsgs[0].Content.StringValue)
}
