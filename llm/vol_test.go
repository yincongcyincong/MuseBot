package llm

import (
	"context"
	"fmt"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestVolSend(t *testing.T) {
	messageChan := make(chan *param.MsgInfo)
	
	go func() {
		for m := range messageChan {
			fmt.Println(m)
		}
	}()
	
	*conf.BaseConfInfo.Type = param.Vol
	
	callLLM := NewLLM(WithChatId("1"), WithMsgId("2"), WithUserId("7"),
		WithMessageChan(messageChan), WithContent("hi"))
	callLLM.LLMClient.GetModel(callLLM)
	callLLM.GetMessages("7", "hi")
	err := callLLM.LLMClient.Send(context.Background(), callLLM)
	assert.Equal(t, nil, err)
}

func TestGetModel(t *testing.T) {
	l := &LLM{UserId: "test-user"}
	h := &VolReq{}
	h.GetModel(l)
	
	assert.Equal(t, param.ModelDeepSeekR1_528, l.Model)
}
