package llm

import (
	"context"
	"errors"
	"testing"
	
	openrouter "github.com/revrost/go-openrouter"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestAIRouterReq_GetModel_Default(t *testing.T) {
	*conf.BaseConfInfo.Type = param.OpenRouter
	mock := &AIRouterReq{}
	l := &LLM{UserId: "non-exist"}
	
	mock.GetModel(l)
	assert.Equal(t, param.DeepseekDeepseekR1_0528Free, l.Model)
}

func TestAIRouterReq_GetModel_UserMode(t *testing.T) {
	
	*conf.BaseConfInfo.Type = param.OpenRouter
	mock := &AIRouterReq{}
	l := &LLM{UserId: "1"}
	mock.GetModel(l)
	assert.Equal(t, param.DeepseekDeepseekR1_0528Free, l.Model)
}

func TestAIRouterReq_GetUserMessage(t *testing.T) {
	mock := &AIRouterReq{}
	mock.GetUserMessage("Hello")
	assert.Equal(t, "user", mock.OpenRouterMsgs[0].Role)
	assert.Equal(t, "Hello", mock.OpenRouterMsgs[0].Content.Text)
}

func TestAIRouterReq_GetAssistantMessage(t *testing.T) {
	mock := &AIRouterReq{}
	mock.GetAssistantMessage("Hi!")
	assert.Equal(t, "assistant", mock.OpenRouterMsgs[0].Role)
	assert.Equal(t, "Hi!", mock.OpenRouterMsgs[0].Content.Text)
}

func TestAIRouterReq_AppendMessages(t *testing.T) {
	main := &AIRouterReq{
		OpenRouterMsgs: []openrouter.ChatCompletionMessage{{Role: "user", Content: openrouter.Content{Text: "Main"}}},
	}
	child := &AIRouterReq{
		OpenRouterMsgs: []openrouter.ChatCompletionMessage{{Role: "assistant", Content: openrouter.Content{Text: "Child"}}},
	}
	
	main.AppendMessages(child)
	assert.Len(t, main.OpenRouterMsgs, 2)
	assert.Equal(t, "Child", main.OpenRouterMsgs[1].Content.Text)
}

func TestAIRouterReq_GetMessage(t *testing.T) {
	mock := &AIRouterReq{}
	mock.GetMessage("user", "test")
	assert.Equal(t, "test", mock.OpenRouterMsgs[0].Content.Text)
	
	mock.GetMessage("assistant", "answer")
	assert.Equal(t, "answer", mock.OpenRouterMsgs[1].Content.Text)
}

func TestAIRouterReq_requestOneToolsCall_JSONError(t *testing.T) {
	r := &AIRouterReq{}
	calls := []openrouter.ToolCall{{
		ID: "call1",
		Function: openrouter.FunctionCall{
			Name:      "fakeTool",
			Arguments: "{invalid json}",
		},
	}}
	// should not panic
	r.requestOneToolsCall(context.Background(), calls, nil)
	assert.Len(t, r.OpenRouterMsgs, 0)
}

func TestAIRouterReq_requestToolsCall_JSONError(t *testing.T) {
	call := openrouter.ChatCompletionStreamChoice{
		Delta: openrouter.ChatCompletionStreamChoiceDelta{
			ToolCalls: []openrouter.ToolCall{{
				ID:   "1",
				Type: "function",
				Function: openrouter.FunctionCall{
					Name:      "invalid",
					Arguments: "{"}}, // malformed JSON
			},
		},
	}
	r := &AIRouterReq{}
	err := r.requestToolsCall(context.Background(), call, nil)
	assert.True(t, errors.Is(err, ToolsJsonErr))
}
