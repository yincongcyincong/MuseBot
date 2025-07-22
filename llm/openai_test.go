package llm

import (
	"context"
	"fmt"
	"os"
	"testing"
	
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

func TestOpenAISend(t *testing.T) {
	messageChan := make(chan *param.MsgInfo)
	
	go func() {
		for m := range messageChan {
			fmt.Println(m)
		}
	}()
	
	*conf.BaseConfInfo.CustomUrl = os.Getenv("TEST_CUSTOM_URL")
	*conf.BaseConfInfo.Type = param.OpenAi
	
	callLLM := NewLLM(WithChatId(1), WithMsgId(2), WithUserId("5"),
		WithMessageChan(messageChan), WithContent("hi"))
	callLLM.LLMClient.GetModel(callLLM)
	callLLM.LLMClient.GetMessages("5", "hi")
	err := callLLM.LLMClient.Send(context.Background(), callLLM)
	assert.Equal(t, nil, err)
	
	*conf.BaseConfInfo.CustomUrl = ""
}

func TestOpenAIReq_GetMessage(t *testing.T) {
	req := &OpenAIReq{}
	req.GetMessage("user", "hello")
	assert.Len(t, req.OpenAIMsgs, 1)
	assert.Equal(t, "hello", req.OpenAIMsgs[0].Content)
	
	req.GetMessage("assistant", "hi")
	assert.Len(t, req.OpenAIMsgs, 2)
	assert.Equal(t, "hi", req.OpenAIMsgs[1].Content)
}

func TestOpenAIReq_AppendMessages(t *testing.T) {
	req1 := &OpenAIReq{}
	req1.GetMessage("user", "message from req1")
	
	req2 := &OpenAIReq{}
	req2.AppendMessages(req1)
	
	assert.Len(t, req2.OpenAIMsgs, 1)
	assert.Equal(t, "message from req1", req2.OpenAIMsgs[0].Content)
}

func TestOpenAIReq_GetModel_Default(t *testing.T) {
	req := &OpenAIReq{}
	llmObj := &LLM{
		UserId: "1",
	}
	
	req.GetModel(llmObj)
	assert.Equal(t, openai.GPT3Dot5Turbo0125, llmObj.Model)
}

func TestRequestToolsCall_InvalidJSON(t *testing.T) {
	req := &OpenAIReq{
		ToolCall: []openai.ToolCall{},
	}
	
	streamChoice := openai.ChatCompletionStreamChoice{
		Delta: openai.ChatCompletionStreamChoiceDelta{
			ToolCalls: []openai.ToolCall{
				{
					ID:   "tool-id",
					Type: "function",
					Function: openai.FunctionCall{
						Name:      "mockTool",
						Arguments: "{invalid-json",
					},
				},
			},
		},
	}
	
	err := req.RequestToolsCall(context.Background(), streamChoice)
	assert.Equal(t, ToolsJsonErr, err)
}
