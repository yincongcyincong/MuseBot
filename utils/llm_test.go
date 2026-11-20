package utils

import (
	"encoding/json"
	"testing"

	godeepseek "github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestGetAvailTxtType_IncludesChatAnyWhere(t *testing.T) {
	conf.BaseConfInfo.ChatAnyWhereToken = "sk-caw-test"
	defer func() { conf.BaseConfInfo.ChatAnyWhereToken = "" }()
	got := GetAvailTxtType()
	assert.Contains(t, got, param.ChatAnyWhere)
}

func TestGetAvailImgType_IncludesChatAnyWhere(t *testing.T) {
	conf.BaseConfInfo.ChatAnyWhereToken = "sk-caw-test"
	defer func() { conf.BaseConfInfo.ChatAnyWhereToken = "" }()
	got := GetAvailImgType()
	assert.Contains(t, got, param.ChatAnyWhere)
}

func TestGetTxtModel_ChatAnyWhereDefault(t *testing.T) {
	// gpt-4o-mini retires on 2027-01-14; gpt-6-luna is the replacement
	assert.Equal(t, param.ModelGPT6Luna, GetTxtModel(param.ChatAnyWhere))
}

func TestGetTxtType_ChatAnyWhereFallback(t *testing.T) {
	conf.BaseConfInfo.Type = param.ChatAnyWhere
	assert.Equal(t, param.ChatAnyWhere, GetTxtType(nil))
}

func TestGetTxtType_TrustsStoredValue(t *testing.T) {
	conf.BaseConfInfo.DeepseekToken = "sk-test"
	conf.BaseConfInfo.Type = param.Gemini
	defer func() {
		conf.BaseConfInfo.DeepseekToken = ""
		conf.BaseConfInfo.Type = ""
	}()

	// the stored type is kept as-is even when its token is not configured
	assert.Equal(t, param.OpenAi, GetTxtType(&param.LLMConfig{TxtType: param.OpenAi}))
	// empty or missing config falls back to the type with a token
	assert.Equal(t, param.DeepSeek, GetTxtType(&param.LLMConfig{}))
	assert.Equal(t, param.DeepSeek, GetTxtType(nil))
}

func TestGetDefaultLLMConfigInjectedPair(t *testing.T) {
	// an injected pair (the last successful conversation from db) is used as-is
	var llmConf param.LLMConfig
	assert.NoError(t, json.Unmarshal([]byte(GetDefaultLLMConfig(param.OpenAi, "glm-5.3-flash")), &llmConf))
	assert.Equal(t, param.OpenAi, llmConf.TxtType)
	assert.Equal(t, "glm-5.3-flash", llmConf.TxtModel)

	// empty pair falls back to the static default chain
	conf.BaseConfInfo.DeepseekToken = "sk-test"
	defer func() { conf.BaseConfInfo.DeepseekToken = "" }()
	llmConf = param.LLMConfig{}
	assert.NoError(t, json.Unmarshal([]byte(GetDefaultLLMConfig("", "")), &llmConf))
	assert.Equal(t, param.DeepSeek, llmConf.TxtType)
	assert.Equal(t, GetTxtModel(param.DeepSeek), llmConf.TxtModel)

	// a type without a model gets the type's static default model
	assert.NoError(t, json.Unmarshal([]byte(GetDefaultLLMConfig(param.Gemini, "")), &llmConf))
	assert.Equal(t, param.Gemini, llmConf.TxtType)
	assert.Equal(t, param.ModelGemini38Flash, llmConf.TxtModel)
}

func TestGetDefaultLLMConfig_UsesAvailTxtType(t *testing.T) {
	conf.BaseConfInfo.DeepseekToken = "sk-test"
	conf.BaseConfInfo.Type = param.Gemini
	defer func() {
		conf.BaseConfInfo.DeepseekToken = ""
		conf.BaseConfInfo.Type = ""
	}()

	var llmConf param.LLMConfig
	assert.NoError(t, json.Unmarshal([]byte(GetDefaultLLMConfig("", "")), &llmConf))
	assert.Equal(t, param.DeepSeek, llmConf.TxtType)
	assert.Equal(t, GetTxtModel(param.DeepSeek), llmConf.TxtModel)
}

func TestGetUsingTxtModel_RelaysKeepAnyModel(t *testing.T) {
	// relay types accept any model, including models natively belonging to
	// another provider
	assert.Equal(t, "deepseek-chat", GetUsingTxtModel(param.OpenAi, godeepseek.DeepSeekChat))
	assert.Equal(t, "my-custom-model", GetUsingTxtModel(param.OpenAi, "my-custom-model"))

	// retired former defaults map to their replacement (#82)
	assert.Equal(t, param.ModelGPT6Luna, GetUsingTxtModel(param.ChatAnyWhere, "gpt-4o-mini"))
	assert.Equal(t, param.ModelGPT6Luna, GetUsingTxtModel(param.OpenAi, "chatgpt-4o-latest"))

	// types with a fixed model list still validate
	assert.Equal(t, param.ModelDeepseekFlash, GetUsingTxtModel(param.DeepSeek, openai.GPT3Dot5Turbo))
	assert.Equal(t, godeepseek.DeepSeekChat, GetUsingTxtModel(param.DeepSeek, godeepseek.DeepSeekChat))
}

func TestGetUsingTTSModel_OpenAIDeprecatedMapping(t *testing.T) {
	// tts-1 family retires on 2027-01-06
	assert.Equal(t, param.ModelGPTRealtime21Mini, GetUsingTTSModel(param.OpenAi, "tts-1"))
	assert.Equal(t, param.ModelGPTRealtime21Mini, GetUsingTTSModel(param.OpenAi, "gpt-4o-mini-tts"))

	// current and custom models pass through
	assert.Equal(t, param.ModelGPTRealtime21Mini, GetUsingTTSModel(param.OpenAi, param.ModelGPTRealtime21Mini))
	assert.Equal(t, "my-custom-tts", GetUsingTTSModel(param.OpenAi, "my-custom-tts"))
}

func TestGetUsingImgModel_OpenAIRetiredMapping(t *testing.T) {
	conf.PhotoConfInfo.OpenAIImageModel = "gpt-image-2.6-sunburst"
	defer func() { conf.PhotoConfInfo.OpenAIImageModel = "" }()

	// retired models stored in user configs map to the replacement (#82)
	assert.Equal(t, param.ModelImageGPT26Sunburst, GetUsingImgModel(param.OpenAi, "gpt-image-1.5"))
	assert.Equal(t, param.ModelImageGPT26Sunburst, GetUsingImgModel(param.OpenAi, "gpt-image-1"))
	assert.Equal(t, param.ModelImageGPT26Sunburst, GetUsingImgModel(param.ChatAnyWhere, "gpt-image-1.5"))

	// current models and custom relays pass through unchanged
	assert.Equal(t, param.ModelImageGPT26Flare, GetUsingImgModel(param.OpenAi, param.ModelImageGPT26Flare))
	assert.Equal(t, "my-custom-model", GetUsingImgModel(param.OpenAi, "my-custom-model"))

	// empty stored model falls back to the configured default
	assert.Equal(t, "gpt-image-2.6-sunburst", GetUsingImgModel(param.OpenAi, ""))
}
