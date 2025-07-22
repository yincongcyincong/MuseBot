// llm/gemini_test.go
package llm_test

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/telegram-deepseek-bot/llm"
	"google.golang.org/genai"
)

func TestGetMessage(t *testing.T) {
	req := &llm.GeminiReq{}
	req.GetUserMessage("hello")
	assert.Equal(t, 1, len(req.GeminiMsgs))
	assert.Equal(t, genai.RoleUser, req.GeminiMsgs[0].Role)
	assert.Equal(t, "hello", req.GeminiMsgs[0].Parts[0].Text)
	
	req.GetAssistantMessage("hi there")
	assert.Equal(t, 2, len(req.GeminiMsgs))
	assert.Equal(t, genai.RoleModel, req.GeminiMsgs[1].Role)
	assert.Equal(t, "hi there", req.GeminiMsgs[1].Parts[0].Text)
}

func TestAppendMessages(t *testing.T) {
	req1 := &llm.GeminiReq{
		GeminiMsgs: []*genai.Content{
			{
				Role:  genai.RoleUser,
				Parts: []*genai.Part{{Text: "A"}},
			},
		},
	}
	req2 := &llm.GeminiReq{
		GeminiMsgs: []*genai.Content{
			{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{{Text: "B"}},
			},
		},
	}
	req1.AppendMessages(req2)
	assert.Equal(t, 2, len(req1.GeminiMsgs))
	assert.Equal(t, "B", req1.GeminiMsgs[1].Parts[0].Text)
}

func TestGenerateGeminiText_EmptyAudio(t *testing.T) {
	text, err := llm.GenerateGeminiText([]byte{})
	assert.Error(t, err)
	assert.Empty(t, text)
}

func TestGenerateGeminiImage_EmptyPrompt(t *testing.T) {
	image, err := llm.GenerateGeminiImg("", nil)
	assert.Error(t, err)
	assert.Nil(t, image)
}

func TestGetGeminiImageContent_EmptyData(t *testing.T) {
	text, err := llm.GetGeminiImageContent([]byte{})
	assert.Error(t, err)
	assert.Empty(t, text)
}

func TestGenerateGeminiVideo_InvalidPrompt(t *testing.T) {
	video, err := llm.GenerateGeminiVideo("")
	assert.Error(t, err)
	assert.Nil(t, video)
}

func TestRequestToolsCall_NilFunctionCall(t *testing.T) {
	req := &llm.GeminiReq{}
	err := req.RequestToolsCall(context.Background(), &genai.GenerateContentResponse{})
	assert.NoError(t, err) // should be a no-op
}

func TestGetModel_DefaultModel(t *testing.T) {
	l := &llm.LLM{}
	req := &llm.GeminiReq{}
	req.GetModel(l)
	assert.NotEmpty(t, l.Model)
}
