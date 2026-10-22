package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
)

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
