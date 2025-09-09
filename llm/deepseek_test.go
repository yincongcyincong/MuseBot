package llm

import (
	"context"
	"os"
	"testing"
	
	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
)

func TestMain(m *testing.M) {
	setup()
	
	code := m.Run()
	
	os.Exit(code)
}

func setup() {
	conf.InitConf()
	db.InitTable()
}

func TestGetMessage_AddsMessageCorrectly(t *testing.T) {
	d := &DeepseekReq{}
	d.GetMessage(constants.ChatMessageRoleUser, "test message")
	
	assert.Len(t, d.DeepseekMsgs, 1)
	assert.Equal(t, constants.ChatMessageRoleUser, d.DeepseekMsgs[0].Role)
	assert.Equal(t, "test message", d.DeepseekMsgs[0].Content)
}

func TestAppendMessages_AppendsCorrectly(t *testing.T) {
	base := &DeepseekReq{
		DeepseekMsgs: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleUser, Content: "first"},
		},
	}
	
	toAppend := &DeepseekReq{
		DeepseekMsgs: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleAssistant, Content: "second"},
		},
	}
	
	base.AppendMessages(toAppend)
	assert.Len(t, base.DeepseekMsgs, 2)
	assert.Equal(t, "second", base.DeepseekMsgs[1].Content)
}

func TestGetAssistantMessage(t *testing.T) {
	d := &DeepseekReq{}
	d.GetAssistantMessage("assistant reply")
	assert.Len(t, d.DeepseekMsgs, 1)
	assert.Equal(t, constants.ChatMessageRoleAssistant, d.DeepseekMsgs[0].Role)
	assert.Equal(t, "assistant reply", d.DeepseekMsgs[0].Content)
}

func TestGetUserMessage(t *testing.T) {
	d := &DeepseekReq{}
	d.GetUserMessage("user message")
	assert.Len(t, d.DeepseekMsgs, 1)
	assert.Equal(t, constants.ChatMessageRoleUser, d.DeepseekMsgs[0].Role)
	assert.Equal(t, "user message", d.DeepseekMsgs[0].Content)
}

func TestRequestToolsCall_JSONError(t *testing.T) {
	d := &DeepseekReq{
		ToolCall: []deepseek.ToolCall{
			{
				Function: deepseek.ToolCallFunction{
					Name:      "mock",
					Arguments: "{invalid-json",
				},
			},
		},
	}
	
	choice := deepseek.StreamChoices{
		Delta: deepseek.StreamDelta{
			ToolCalls: d.ToolCall,
		},
	}
	
	err := d.RequestToolsCall(context.TODO(), choice, nil)
	assert.Equal(t, ToolsJsonErr, err)
}

func TestGetMessage_AppendsCorrectlyWhenNotEmpty(t *testing.T) {
	d := &DeepseekReq{
		DeepseekMsgs: []deepseek.ChatCompletionMessage{
			{Role: constants.ChatMessageRoleUser, Content: "prev"},
		},
	}
	
	d.GetMessage(constants.ChatMessageRoleAssistant, "next")
	
	assert.Len(t, d.DeepseekMsgs, 2)
	assert.Equal(t, "next", d.DeepseekMsgs[1].Content)
	assert.Equal(t, constants.ChatMessageRoleAssistant, d.DeepseekMsgs[1].Role)
}
