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
	assert.Equal(t, openai.GPT3Dot5Turbo, GetTxtModel(param.ChatAnyWhere))
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

func TestGetDefaultLLMConfig_UsesAvailTxtType(t *testing.T) {
	conf.BaseConfInfo.DeepseekToken = "sk-test"
	conf.BaseConfInfo.Type = param.Gemini
	defer func() {
		conf.BaseConfInfo.DeepseekToken = ""
		conf.BaseConfInfo.Type = ""
	}()

	var llmConf param.LLMConfig
	assert.NoError(t, json.Unmarshal([]byte(GetDefaultLLMConfig()), &llmConf))
	assert.Equal(t, param.DeepSeek, llmConf.TxtType)
	assert.Equal(t, GetTxtModel(param.DeepSeek), llmConf.TxtModel)
}

func TestGetUsingTxtModel_RelaysKeepAnyModel(t *testing.T) {
	// relay types accept any model, including models natively belonging to
	// another provider
	assert.Equal(t, "deepseek-chat", GetUsingTxtModel(param.OpenAi, godeepseek.DeepSeekChat))
	assert.Equal(t, "my-custom-model", GetUsingTxtModel(param.OpenAi, "my-custom-model"))

	// types with a fixed model list still validate
	assert.Equal(t, godeepseek.DeepSeekChat, GetUsingTxtModel(param.DeepSeek, openai.GPT3Dot5Turbo))
	assert.Equal(t, godeepseek.DeepSeekChat, GetUsingTxtModel(param.DeepSeek, godeepseek.DeepSeekChat))
}
