package llm

import (
	"context"
	"testing"
	
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

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
