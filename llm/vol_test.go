package llm

import (
	"context"
	"fmt"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
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
	
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"vol"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Vol},
	})
	
	callLLM := NewLLM(WithChatId("1"), WithMsgId("2"), WithUserId("7"),
		WithMessageChan(messageChan), WithContent("hi"), WithContext(ctx))
	callLLM.LLMClient.GetModel(callLLM)
	callLLM.GetMessages("7", "hi")
	err := callLLM.LLMClient.Send(ctx, callLLM)
	assert.Equal(t, nil, err)
}
