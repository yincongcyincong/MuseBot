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

func TestGeminiSend(t *testing.T) {
	conf.InitConf()
	messageChan := make(chan *param.MsgInfo)
	
	go func() {
		for m := range messageChan {
			fmt.Println(m)
		}
	}()
	
	conf.BaseConfInfo.Type = param.Gemini
	
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	callLLM := NewLLM(WithChatId("1"), WithMsgId("2"), WithUserId("4"),
		WithMessageChan(messageChan), WithContent("hi"), WithContext(ctx))
	callLLM.LLMClient.GetModel(callLLM)
	callLLM.GetMessages("4", "hi")
	err := callLLM.LLMClient.Send(ctx, callLLM)
	assert.Equal(t, nil, err)
	
}

func TestGenerateGeminiText_EmptyAudio(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	text, _, err := GenerateGeminiText(ctx, []byte{})
	assert.Error(t, err)
	assert.Empty(t, text)
}

func TestGenerateGeminiImage_EmptyPrompt(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	image, _, err := GenerateGeminiImg(ctx, "", nil)
	assert.Error(t, err)
	assert.Nil(t, image)
}

func TestGenerateGeminiVideo_InvalidPrompt(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_info", &db.User{
		LLMConfig:    `{"type":"gemini"}`,
		LLMConfigRaw: &param.LLMConfig{TxtType: param.Gemini},
	})
	video, _, err := GenerateGeminiVideo(ctx, "", nil)
	assert.Error(t, err)
	assert.Nil(t, video)
}
