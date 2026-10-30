package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/param"
	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
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

func TestInteractionStepError(t *testing.T) {
	// nil response and empty steps carry no error
	assert.NoError(t, interactionStepError(nil))
	assert.NoError(t, interactionStepError(&operations.CreateInteractionResponse{}))
	assert.NoError(t, interactionStepError(&operations.CreateInteractionResponse{
		Interaction: &interactions.Interaction{},
	}))

	// a failed model-output step surfaces the server's message
	res := &operations.CreateInteractionResponse{
		Interaction: &interactions.Interaction{Steps: []interactions.Step{{
			ModelOutputStep: &interactions.ModelOutputStep{
				Error: &interactions.Status{Message: genai.Ptr("quota exceeded")},
			},
		}}},
	}
	err := interactionStepError(res)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "quota exceeded")
	}

	// an error step without a message still fails explicitly
	res.Interaction.Steps[0].ModelOutputStep.Error.Message = nil
	err = interactionStepError(res)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "generation failed")
	}
}

func TestInteractionID(t *testing.T) {
	assert.Equal(t, "", interactionID(nil))
	assert.Equal(t, "", interactionID(&operations.CreateInteractionResponse{}))
	assert.Equal(t, "", interactionID(&operations.CreateInteractionResponse{
		Interaction: &interactions.Interaction{},
	}))
	assert.Equal(t, "abc", interactionID(&operations.CreateInteractionResponse{
		Interaction: &interactions.Interaction{ID: genai.Ptr("abc")},
	}))
}
