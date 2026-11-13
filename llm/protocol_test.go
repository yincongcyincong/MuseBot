package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestLLMToolStateAccumulatesStreamedCall(t *testing.T) {
	tools := &LLMToolState{}

	// sequential provider: fragments carry no index (-1) and merge in wire
	// order into the most recent call
	assert.ErrorIs(t, tools.observe(LLMToolCall{
		Index:     -1,
		ID:        "call-1",
		Name:      "web_search",
		Type:      "function",
		Arguments: `{"quer`,
	}), ToolsJsonErr)
	assert.Empty(t, tools.Results)

	assert.NoError(t, tools.observe(LLMToolCall{Index: -1, Arguments: `y":"mcp"} `}))
	assert.Len(t, tools.Calls, 1)
	assert.Equal(t, "call-1", tools.Calls[0].ID)
	assert.Equal(t, "web_search", tools.Calls[0].Name)
	assert.Equal(t, `{"query":"mcp"} `, tools.Calls[0].Arguments)
	assert.Empty(t, tools.Results)
}

func TestLLMToolStateRoutesInterleavedToolCallFragments(t *testing.T) {
	tools := &LLMToolState{}

	// parallel calls arrive as interleaved fragments tagged with the call
	// index (OpenAI streaming spec); they must stay separate
	assert.ErrorIs(t, tools.observe(LLMToolCall{
		Index: 0, ID: "call-1", Name: "web_search", Type: "function",
		Arguments: `{"quer`,
	}), ToolsJsonErr)
	assert.ErrorIs(t, tools.observe(LLMToolCall{
		Index: 1, ID: "call-2", Name: "shell", Type: "function",
		Arguments: `{"ar`,
	}), ToolsJsonErr)

	assert.NoError(t, tools.observe(LLMToolCall{Index: 1, Arguments: `gs":[1]}`}))
	assert.NoError(t, tools.observe(LLMToolCall{Index: 0, Arguments: `y":"mcp"}`}))

	assert.Len(t, tools.Calls, 2)
	assert.Equal(t, "web_search", tools.Calls[0].Name)
	assert.Equal(t, `{"query":"mcp"}`, tools.Calls[0].Arguments)
	assert.Equal(t, "shell", tools.Calls[1].Name)
	assert.Equal(t, `{"args":[1]}`, tools.Calls[1].Arguments)
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

func TestSanitizeOpenAISamplingParams(t *testing.T) {
	// reasoning-era models only accept default sampling values and moved
	// max_tokens to max_completion_tokens
	request := openai.ChatCompletionRequest{
		Model:            param.ModelGPT52,
		Temperature:      0.7,
		TopP:             0.9,
		PresencePenalty:  0.5,
		FrequencyPenalty: 0.5,
		LogProbs:         true,
		TopLogProbs:      3,
		MaxTokens:        512,
		Stop:             []string{"stop"},
	}
	sanitizeOpenAISamplingParams(&request)

	assert.Zero(t, request.Temperature)
	assert.Zero(t, request.TopP)
	assert.Zero(t, request.PresencePenalty)
	assert.Zero(t, request.FrequencyPenalty)
	assert.False(t, request.LogProbs)
	assert.Zero(t, request.TopLogProbs)
	assert.Zero(t, request.MaxTokens)
	assert.Equal(t, 512, request.MaxCompletionTokens)
	// stop survives: still supported on gpt-5
	assert.Equal(t, []string{"stop"}, request.Stop)

	// every gpt-5 / o-series family is treated as reasoning
	for _, model := range []string{
		param.ModelGPT52Pro, param.ModelGPT5Mini, "o1", "o1-mini", "o3", "o4-mini",
	} {
		assert.True(t, isOpenAIReasoningModel(model), model)
	}

	// classic models and other providers routed through the OpenAI client
	// keep the configured sampling options untouched
	request = openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		Temperature: 0.7,
		MaxTokens:   512,
	}
	sanitizeOpenAISamplingParams(&request)
	assert.Equal(t, float32(0.7), request.Temperature)
	assert.Equal(t, 512, request.MaxTokens)
	assert.Zero(t, request.MaxCompletionTokens)

	request = openai.ChatCompletionRequest{Model: param.ModelGemini38Flash, Temperature: 0.7}
	sanitizeOpenAISamplingParams(&request)
	assert.Equal(t, float32(0.7), request.Temperature)
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

func TestParseXMLToolCalls(t *testing.T) {
	// canonical GLM JSON form
	content := "我来查询当前时间\n<tool_call>\n{\"name\": \"mcp__time__get_current_time\", \"arguments\": {\"timezone\": \"Asia/Shanghai\"}}\n</tool_call>\n"
	calls, remaining := parseXMLToolCalls(content)
	require.Len(t, calls, 1)
	assert.Equal(t, "mcp__time__get_current_time", calls[0].Name)
	assert.NotEmpty(t, calls[0].ID)
	assert.Equal(t, "function", calls[0].Type)
	assert.JSONEq(t, `{"timezone": "Asia/Shanghai"}`, calls[0].Arguments)
	assert.NotContains(t, remaining, "tool_call")
	assert.Contains(t, remaining, "我来查询当前时间")

	// loose form: bare name on its own line, JSON below (what glm-5.3-flash
	// actually emitted in the wild)
	content = "<tool_call>\nmcp__time__get_current_time\n{\"timezone\": \"Asia/Shanghai\"}\n</tool_call>"
	calls, remaining = parseXMLToolCalls(content)
	require.Len(t, calls, 1)
	assert.Equal(t, "mcp__time__get_current_time", calls[0].Name)
	assert.JSONEq(t, `{"timezone": "Asia/Shanghai"}`, calls[0].Arguments)
	assert.Empty(t, strings.TrimSpace(remaining))

	// several blocks, ids stay unique
	content = `<tool_call>{"name": "a"}</tool_call>text<tool_call>{"name": "b"}</tool_call>`
	calls, remaining = parseXMLToolCalls(content)
	require.Len(t, calls, 2)
	assert.Equal(t, "a", calls[0].Name)
	assert.Equal(t, "{}", calls[0].Arguments)
	assert.Equal(t, "b", calls[1].Name)
	assert.NotEqual(t, calls[0].ID, calls[1].ID)
	assert.Equal(t, "text", remaining)

	// no arguments at all, and a no-name block is skipped
	_, remaining = parseXMLToolCalls("<tool_call>\n{\"arguments\": {}}\n</tool_call>")
	assert.Contains(t, remaining, "tool_call", "unparseable blocks stay in the content")

	// plain content without blocks is untouched
	content = "现在几点了？没有工具调用。"
	calls, remaining = parseXMLToolCalls(content)
	assert.Nil(t, calls)
	assert.Equal(t, content, remaining)
}

func TestAdoptInlineToolCalls(t *testing.T) {
	l := &LLM{Ctx: context.Background()}
	l.WholeContent = "前置说明\n<tool_call>\n{\"name\": \"web_search\", \"arguments\": {\"query\": \"q\"}}\n</tool_call>"
	tools := &LLMToolState{}
	msgInfo := &param.MsgInfo{Content: "<tool_call>\n{\"name\": \"web_search\", \"arguments\": {\"query\": \"q\"}}\n</tool_call>"}

	assert.True(t, adoptInlineToolCalls(l, tools, msgInfo))
	require.Len(t, tools.Calls, 1)
	assert.Equal(t, "web_search", tools.Calls[0].Name)
	assert.NotContains(t, l.WholeContent, "tool_call")
	assert.Contains(t, l.WholeContent, "前置说明")
	assert.NotContains(t, msgInfo.Content, "tool_call")

	// second adoption on plain content is a no-op
	assert.False(t, adoptInlineToolCalls(l, tools, msgInfo))
	assert.Len(t, tools.Calls, 1)
}
