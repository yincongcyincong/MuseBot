package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

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
	value, err := retryLLMRequest(nil, "sync", func() (string, error) {
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

	// every gpt-5 / gpt-6 / o-series family is treated as reasoning
	for _, model := range []string{
		param.ModelGPT52Pro, param.ModelGPT5Mini, param.ModelGPT6Luna, "o1", "o1-mini", "o3", "o4-mini",
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

func TestAdoptInlineToolBlocks(t *testing.T) {
	l := &LLM{Ctx: context.Background()}
	tools := &LLMToolState{}

	assert.True(t, adoptInlineToolBlocks(l, tools, []string{
		"<tool_call>\n{\"name\": \"web_search\", \"arguments\": {\"query\": \"q\"}}\n</tool_call>",
	}))
	require.Len(t, tools.Calls, 1)
	assert.Equal(t, "web_search", tools.Calls[0].Name)
	assert.JSONEq(t, `{"query": "q"}`, tools.Calls[0].Arguments)

	// nothing new to adopt is a no-op
	assert.False(t, adoptInlineToolBlocks(l, tools, nil))
	assert.Len(t, tools.Calls, 1)
}

// TestXMLToolCallGateWithholdsBlocks drives the streaming gate with the chunk
// fragmentation a real stream produces (tag split across chunks) and asserts
// no raw block text is ever forwarded.
func TestXMLToolCallGateWithholdsBlocks(t *testing.T) {
	var g xmlToolCallGate
	var forwarded strings.Builder

	for _, chunk := range []string{
		"我来查询当前时间\n<tool_ca",
		"ll>\nmcp__time__get_current_time\n{\"timezone\": \"Asia/Shanghai\"}\n</tool_",
		"call>\n以上是查询计划。",
	} {
		forwarded.WriteString(g.push(chunk))
	}
	rest, blocks := g.finish()
	forwarded.WriteString(rest)

	assert.Equal(t, "我来查询当前时间\n\n以上是查询计划。", forwarded.String())
	require.Len(t, blocks, 1)
	assert.Contains(t, blocks[0], "mcp__time__get_current_time")
}

func TestXMLToolCallGateFalseAlarmFlushes(t *testing.T) {
	var g xmlToolCallGate
	var forwarded strings.Builder
	for _, chunk := range []string{"对比 a<b 和 ", "<toolbar", "> 的用法", "，还有 <tool", "ate 的例子"} {
		forwarded.WriteString(g.push(chunk))
	}
	rest, blocks := g.finish()
	forwarded.WriteString(rest)

	// partial "<tool" turned out not to be the tag: everything flushes
	assert.Equal(t, "对比 a<b 和 <toolbar> 的用法，还有 <toolate 的例子", forwarded.String())
	assert.Empty(t, blocks)
}

// TestXMLToolCallGateDropsUnterminatedBlock: a block the stream never closed
// is dropped at finish instead of leaking its raw XML.
func TestXMLToolCallGateDropsUnterminatedBlock(t *testing.T) {
	var g xmlToolCallGate
	forwarded := g.push("开头说明 <tool_call>{\"name\": \"x\"")
	rest, blocks := g.finish()
	forwarded += rest

	assert.Equal(t, "开头说明 ", forwarded)
	assert.Empty(t, blocks)
}

// TestReceiveLLMStreamInlineToolCall runs the full streaming path: content
// with an embedded tool-call block streams to the user without the raw XML
// and the block executes as a real tool call.
func TestReceiveLLMStreamInlineToolCall(t *testing.T) {
	previous := conf.ToolsConfInfo.WebSearchProvider
	conf.ToolsConfInfo.WebSearchProvider = "disabled" // offline: the call fails and is fed back
	t.Cleanup(func() { conf.ToolsConfInfo.WebSearchProvider = previous })

	l := &LLM{Ctx: context.Background(), Model: "test"}
	tools := &LLMToolState{}
	chunks := []string{
		"先查一下。\n<tool_call>\n",
		"{\"name\": \"web_search\", \"arguments\": {\"query\": \"q\"}}\n</tool_call>\n",
	}
	next := func() (LLMStreamChunk, error) {
		if len(chunks) == 0 {
			return LLMStreamChunk{}, io.EOF
		}
		chunk := LLMStreamChunk{Content: chunks[0]}
		chunks = chunks[1:]
		return chunk, nil
	}

	msgInfo, hasTools, err := receiveLLMStream(context.Background(), l, time.Now(), next, tools)
	require.NoError(t, err)
	assert.True(t, hasTools)
	require.Len(t, tools.Calls, 1)
	assert.Equal(t, "web_search", tools.Calls[0].Name)
	// neither the replayed content nor the streamed message carries the XML
	assert.Equal(t, "先查一下。\n\n", l.WholeContent) // preamble + the newline after the block
	if msgInfo != nil {
		assert.NotContains(t, msgInfo.Content, "tool_call")
	}
	// the call ran (disabled provider surfaces as the fed-back failure)
	require.Len(t, tools.Results, 1)
	assert.Contains(t, tools.Results[0].Content, "web_search")
}
