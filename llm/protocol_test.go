package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
)

func TestLLMToolStateAccumulatesStreamedCall(t *testing.T) {
	tools := &LLMToolState{}

	assert.ErrorIs(t, tools.observe(LLMToolCall{
		ID:        "call-1",
		Name:      "web_search",
		Type:      "function",
		Arguments: `{"quer`,
	}), ToolsJsonErr)
	assert.Empty(t, tools.Results)

	assert.NoError(t, tools.observe(LLMToolCall{Arguments: `y":"mcp"} `}))
	assert.Len(t, tools.Calls, 1)
	assert.Equal(t, "call-1", tools.Calls[0].ID)
	assert.Equal(t, "web_search", tools.Calls[0].Name)
	assert.Equal(t, `{"query":"mcp"} `, tools.Calls[0].Arguments)
	assert.Empty(t, tools.Results)
}

func TestRetryLLMRequestRetriesUntilSuccess(t *testing.T) {
	previousTimes := conf.BaseConfInfo.LLMRetryTimes
	previousInterval := conf.BaseConfInfo.LLMRetryInterval
	conf.BaseConfInfo.LLMRetryTimes = 2
	conf.BaseConfInfo.LLMRetryInterval = 0
	t.Cleanup(func() {
		conf.BaseConfInfo.LLMRetryTimes = previousTimes
		conf.BaseConfInfo.LLMRetryInterval = previousInterval
	})

	attempts := 0
	value, err := retryLLMRequest(nil, func() (string, error) {
		attempts++
		if attempts == 1 {
			return "", assert.AnError
		}
		return "ok", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "ok", value)
	assert.Equal(t, 2, attempts)
}

func TestLLMSamplingOptionsApplyToProviderRequests(t *testing.T) {
	previousOption := conf.BaseConfInfo.LLMOptionParam
	previousLLMConf := *conf.LLMConfInfo
	conf.BaseConfInfo.LLMOptionParam = true
	conf.LLMConfInfo.MaxTokens = 128
	conf.LLMConfInfo.TopP = 0.5
	conf.LLMConfInfo.FrequencyPenalty = 0.25
	conf.LLMConfInfo.TopLogProbs = 3
	conf.LLMConfInfo.LogProbs = true
	conf.LLMConfInfo.Stop = []string{"stop"}
	conf.LLMConfInfo.PresencePenalty = 0.75
	conf.LLMConfInfo.Temperature = 0.2
	t.Cleanup(func() {
		conf.BaseConfInfo.LLMOptionParam = previousOption
		*conf.LLMConfInfo = previousLLMConf
	})

	openAIRequest := &openai.ChatCompletionRequest{}
	currentLLMSamplingOptions().applyToOpenAI(openAIRequest)
	deepseekRequest := &deepseek.ChatCompletionRequest{}
	currentLLMSamplingOptions().applyToDeepseek(deepseekRequest)
	deepseekStreamRequest := &deepseek.StreamChatCompletionRequest{}
	currentLLMSamplingOptions().applyToDeepseekStream(deepseekStreamRequest)

	for _, request := range []struct {
		maxTokens        int
		topLogProbs      int
		frequencyPenalty float32
		presencePenalty  float32
		temperature      float32
		topP             float32
	}{
		{
			maxTokens:        openAIRequest.MaxTokens,
			topLogProbs:      openAIRequest.TopLogProbs,
			frequencyPenalty: openAIRequest.FrequencyPenalty,
			presencePenalty:  openAIRequest.PresencePenalty,
			temperature:      openAIRequest.Temperature,
			topP:             openAIRequest.TopP,
		},
		{
			maxTokens:        deepseekRequest.MaxTokens,
			topLogProbs:      deepseekRequest.TopLogProbs,
			frequencyPenalty: deepseekRequest.FrequencyPenalty,
			presencePenalty:  deepseekRequest.PresencePenalty,
			temperature:      deepseekRequest.Temperature,
			topP:             deepseekRequest.TopP,
		},
		{
			maxTokens:        deepseekStreamRequest.MaxTokens,
			topLogProbs:      deepseekStreamRequest.TopLogProbs,
			frequencyPenalty: deepseekStreamRequest.FrequencyPenalty,
			presencePenalty:  deepseekStreamRequest.PresencePenalty,
			temperature:      deepseekStreamRequest.Temperature,
			topP:             deepseekStreamRequest.TopP,
		},
	} {
		assert.Equal(t, 128, request.maxTokens)
		assert.Equal(t, 3, request.topLogProbs)
		assert.Equal(t, float32(0.25), request.frequencyPenalty)
		assert.Equal(t, float32(0.75), request.presencePenalty)
		assert.Equal(t, float32(0.2), request.temperature)
		assert.Equal(t, float32(0.5), request.topP)
	}

	assert.True(t, openAIRequest.LogProbs)
	assert.True(t, deepseekRequest.LogProbs)
	assert.True(t, deepseekStreamRequest.LogProbs)
	assert.Equal(t, conf.LLMConfInfo.Stop, openAIRequest.Stop)
	assert.Equal(t, conf.LLMConfInfo.Stop, deepseekRequest.Stop)
	assert.Equal(t, conf.LLMConfInfo.Stop, deepseekStreamRequest.Stop)
}

func TestOllamaClientMessageMutation(t *testing.T) {
	var client LLMClient = &OllamaReq{}
	client.GetMessage("user", "hello")

	req := client.(*OllamaReq)
	assert.Len(t, req.OllamaMsgs, 1)
	assert.Equal(t, "hello", req.OllamaMsgs[0].Content)
}

func TestLLMToolStateBlocksRepeatedFailures(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "disabled"
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })

	l := searchTaskAgent(newSearchTask())
	args, _ := json.Marshal(map[string]string{"query": "q"})
	state := &LLMToolState{}

	for i := 0; i < taskToolMaxConsecutiveFails+1; i++ {
		state.Calls = []LLMToolCall{{
			ID:        fmt.Sprintf("call-%d", i),
			Name:      conf.BuiltinWebSearchTool,
			Arguments: string(args),
		}}
		assert.NoError(t, state.execute(context.Background(), l))
	}

	assert.Len(t, state.Results, taskToolMaxConsecutiveFails+1)
	for i, result := range state.Results {
		if i < taskToolMaxConsecutiveFails {
			assert.Contains(t, result.Content, "failed")
		} else {
			assert.Contains(t, result.Content, "blocked")
		}
	}
}
