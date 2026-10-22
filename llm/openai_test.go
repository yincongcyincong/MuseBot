package llm

import (
	"testing"

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

func TestRequestToolsCall_InvalidJSON(t *testing.T) {
	tools := &LLMToolState{}

	err := tools.observe(LLMToolCall{
		ID:        "tool-id",
		Name:      "mockTool",
		Type:      "function",
		Arguments: "{invalid-json",
	})
	assert.Equal(t, ToolsJsonErr, err)
}
