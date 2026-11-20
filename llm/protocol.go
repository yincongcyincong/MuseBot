package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/cohesion-org/deepseek-go"
	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
)

type LLMToolCall struct {
	// Index routes streamed continuation fragments back to their call when
	// the provider supplies one (OpenAI streaming spec). -1 = not provided.
	Index     int
	ID        string
	Name      string
	Type      string
	Arguments string
}

type LLMToolResult struct {
	ID      string
	Content string
}

type LLMToolState struct {
	Calls   []LLMToolCall
	Results []LLMToolResult

	// circuit breaker: stop executing a tool that keeps failing with the same
	// request, so a stubborn model cannot burn the whole loop budget on it
	lastFailTool     string
	consecutiveFails int
	blockedTool      string
}

type LLMStreamChunk struct {
	Content   string
	ToolCalls []LLMToolCall
	Usage     int
	HasUsage  bool
}

// retryRequest reruns an API call up to LLMRetryTimes, sleeping
// LLMRetryInterval between attempts (never after the final one) and giving up
// early once ctx is cancelled. LLMRetryTimes <= 0 means a single attempt.
func retryRequest[T any](ctx context.Context, operation func() (T, error)) (T, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var value T
	var err error

	times := conf.BaseConfInfo.LLMRetryTimes
	if times <= 0 {
		times = 1
	}
	for attempt := 0; attempt < times; attempt++ {
		value, err = operation()
		if err == nil {
			return value, nil
		}
		if attempt == times-1 {
			break
		}
		select {
		case <-ctx.Done():
			return value, ctx.Err()
		case <-time.After(time.Duration(conf.BaseConfInfo.LLMRetryInterval) * time.Millisecond):
		}
	}
	return value, err
}

func retryLLMRequest[T any](l *LLM, mode string, operation func() (T, error)) (T, error) {
	if l == nil {
		return retryRequest(nil, operation)
	}
	provider, model := l.txtProviderType(), l.Model
	return retryRequest(l.Ctx, func() (T, error) {
		start := time.Now()
		value, err := operation()
		outcome := metrics.Outcome(l.Ctx, err)
		metrics.LLMAttempts.WithLabelValues(provider, model, mode, outcome).Inc()
		metrics.LLMAttemptDuration.WithLabelValues(provider, model, mode, outcome).Observe(time.Since(start).Seconds())
		return value, err
	})
}

type llmSamplingOptions struct {
	FrequencyPenalty float32
	MaxTokens        int
	PresencePenalty  float32
	Temperature      float32
	TopLogProbs      int
	TopP             float32
	Stop             []string
	LogProbs         bool
}

func currentLLMSamplingOptions() llmSamplingOptions {
	if !conf.BaseConfInfo.LLMOptionParam {
		return llmSamplingOptions{}
	}

	return llmSamplingOptions{
		FrequencyPenalty: float32(conf.LLMConfInfo.FrequencyPenalty),
		MaxTokens:        conf.LLMConfInfo.MaxTokens,
		PresencePenalty:  float32(conf.LLMConfInfo.PresencePenalty),
		Temperature:      float32(conf.LLMConfInfo.Temperature),
		TopLogProbs:      conf.LLMConfInfo.TopLogProbs,
		TopP:             float32(conf.LLMConfInfo.TopP),
		Stop:             conf.LLMConfInfo.Stop,
		LogProbs:         conf.LLMConfInfo.LogProbs,
	}
}

func (o llmSamplingOptions) applyToOpenAI(request *openai.ChatCompletionRequest) {
	request.FrequencyPenalty = o.FrequencyPenalty
	request.MaxTokens = o.MaxTokens
	request.PresencePenalty = o.PresencePenalty
	request.Temperature = o.Temperature
	request.TopLogProbs = o.TopLogProbs
	request.TopP = o.TopP
	request.Stop = o.Stop
	request.LogProbs = o.LogProbs
}

// isOpenAIReasoningModel reports whether the model belongs to the
// reasoning-era OpenAI families (gpt-5.x, gpt-6.x, o1/o3/o4) that only
// accept the default sampling values and reject logprobs.
func isOpenAIReasoningModel(model string) bool {
	return strings.HasPrefix(model, "gpt-5") ||
		strings.HasPrefix(model, "gpt-6") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}

// sanitizeOpenAISamplingParams drops parameters the reasoning-era OpenAI
// models reject: temperature/top_p/penalties may only stay at their default,
// logprobs are unsupported, and max_tokens was replaced by
// max_completion_tokens. All fields are omitempty, so zeroing them omits the
// parameter instead of sending an invalid value.
func sanitizeOpenAISamplingParams(request *openai.ChatCompletionRequest) {
	if !isOpenAIReasoningModel(request.Model) {
		return
	}

	request.Temperature = 0
	request.TopP = 0
	request.PresencePenalty = 0
	request.FrequencyPenalty = 0
	request.LogProbs = false
	request.TopLogProbs = 0
	if request.MaxTokens > 0 {
		request.MaxCompletionTokens = request.MaxTokens
		request.MaxTokens = 0
	}
}

func (o llmSamplingOptions) applyToDeepseek(request *deepseek.ChatCompletionRequest) {
	request.FrequencyPenalty = o.FrequencyPenalty
	request.MaxTokens = o.MaxTokens
	request.PresencePenalty = o.PresencePenalty
	request.Temperature = o.Temperature
	request.TopLogProbs = o.TopLogProbs
	request.TopP = o.TopP
	request.Stop = o.Stop
	request.LogProbs = o.LogProbs
}

func (o llmSamplingOptions) applyToDeepseekStream(request *deepseek.StreamChatCompletionRequest) {
	request.FrequencyPenalty = o.FrequencyPenalty
	request.MaxTokens = o.MaxTokens
	request.PresencePenalty = o.PresencePenalty
	request.Temperature = o.Temperature
	request.TopLogProbs = o.TopLogProbs
	request.TopP = o.TopP
	request.Stop = o.Stop
	request.LogProbs = o.LogProbs
}

// observe accumulates one streamed tool-call fragment. The first fragment of
// a call carries id+name (and, on providers following the OpenAI streaming
// spec, an index); continuation fragments carry arguments only and are routed
// back to their call by index, so parallel tool calls with interleaved
// argument fragments stay separate.
func (s *LLMToolState) observe(call LLMToolCall) error {
	target := -1
	if call.Name != "" || len(s.Calls) == 0 {
		// start of a new tool call
		s.Calls = append(s.Calls, call)
		target = len(s.Calls) - 1
	} else {
		target = s.fragmentTarget(call.Index)
		if call.ID != "" {
			s.Calls[target].ID = call.ID
		}
		if call.Type != "" {
			s.Calls[target].Type = call.Type
		}
		s.Calls[target].Arguments += call.Arguments
	}

	var property map[string]interface{}
	if err := json.Unmarshal([]byte(s.Calls[target].Arguments), &property); err != nil {
		return ToolsJsonErr
	}
	return nil
}

// fragmentTarget locates the call a continuation fragment belongs to. An
// index that maps to exactly one recorded call routes there; duplicated or
// missing indexes (sequential providers that omit the field) fall back to the
// most recent call, which matches those providers' wire order.
func (s *LLMToolState) fragmentTarget(index int) int {
	if index >= 0 {
		found, matches := -1, 0
		for i := range s.Calls {
			if s.Calls[i].Index == index {
				found = i
				matches++
			}
		}
		if matches == 1 {
			return found
		}
	}
	return len(s.Calls) - 1
}

// taskToolMaxConsecutiveFails bounds how many times in a row the same tool
// may fail before it is blocked for the rest of the request
const taskToolMaxConsecutiveFails = 3

func (s *LLMToolState) execute(ctx context.Context, l *LLM) error {
	for _, call := range s.Calls {
		if s.blockedTool == call.Name {
			logger.WarnCtx(l.Ctx, "tool call skipped, blocked after repeated failures", "tool", call.Name)
			s.Results = append(s.Results, LLMToolResult{
				ID:      call.ID,
				Content: fmt.Sprintf("tool %s is blocked: it already failed %d times in a row. Do not call it again; use a different approach or produce your final answer with the information collected so far.", call.Name, s.consecutiveFails),
			})
			continue
		}

		var property map[string]interface{}
		if err := json.Unmarshal([]byte(call.Arguments), &property); err != nil {
			logger.WarnCtx(l.Ctx, "tool arguments json unmarshal fail", "tool", call.Name, "args", call.Arguments, "err", err)
			// invalid arguments are fed back as the tool result so the LLM can
			// correct them instead of aborting the whole request
			l.toolCalls++
			l.toolFails++
			s.Results = append(s.Results, LLMToolResult{
				ID:      call.ID,
				Content: fmt.Sprintf("tool %s failed: invalid arguments: %v", call.Name, err),
			})
			continue
		}

		toolStart := time.Now()
		content, err := l.ExecMcpReq(ctx, call.Name, property)
		// nil-safe: reports to the admin flow view only inside /task runs
		l.taskTracker.ToolCalled(call.Name, call.Arguments, err, time.Since(toolStart))
		l.toolCalls++
		if err != nil {
			// a dead context means the request itself is being torn down, so
			// there is no point in feeding the failure back to the LLM
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			logger.ErrorCtx(ctx, "execute tool fail", "toolId", call.ID, "tool", call.Name, "err", err)
			l.toolFails++
			// tool failures are fed back as the tool result so the LLM can adapt
			// (different command, tool, or answer in the summary); only LLM
			// request errors abort the request
			if s.lastFailTool == call.Name {
				s.consecutiveFails++
			} else {
				s.lastFailTool = call.Name
				s.consecutiveFails = 1
			}
			if s.consecutiveFails >= taskToolMaxConsecutiveFails {
				s.blockedTool = call.Name
			}
			s.Results = append(s.Results, LLMToolResult{
				ID:      call.ID,
				Content: fmt.Sprintf("tool %s failed: %v", call.Name, err),
			})
			continue
		}
		s.lastFailTool = ""
		s.consecutiveFails = 0
		s.Results = append(s.Results, LLMToolResult{ID: call.ID, Content: content})
	}
	return nil
}

func (s *LLMToolState) observeComplete(ctx context.Context, l *LLM, calls []LLMToolCall) error {
	s.Calls = append(s.Calls, calls...)
	return s.execute(ctx, l)
}

func (s *LLMToolState) reset() {
	s.Calls = nil
	s.Results = nil
}

// xmlToolCallRe matches one <tool_call>...</tool_call> block. GLM-style
// models emit their native agentic tool-call syntax in the message content
// instead of the native function-call fields. The tag spelling is strict
// lowercase so the streaming gate and this parser agree on the same blocks.
var xmlToolCallRe = regexp.MustCompile(`(?s)<tool_call>\s*(.*?)\s*</tool_call>`)

const (
	xmlToolCallOpen  = "<tool_call>"
	xmlToolCallClose = "</tool_call>"
)

// parseXMLToolCallInner parses the inside of one <tool_call> block. Two
// shapes are accepted:
//   - the canonical GLM JSON form: {"name": "...", "arguments": {...}}
//   - a loose form with the bare name on the first line and the JSON
//     arguments on the following line(s)
func parseXMLToolCallInner(inner string, idx int) (LLMToolCall, bool) {
	inner = strings.TrimSpace(inner)
	call := LLMToolCall{Index: -1, ID: fmt.Sprintf("xmltool_%d", idx), Type: "function"}
	if strings.HasPrefix(inner, "{") {
		var parsed struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(inner), &parsed); err == nil {
			call.Name = parsed.Name
			call.Arguments = string(parsed.Arguments)
		}
	} else {
		lines := strings.SplitN(inner, "\n", 2)
		call.Name = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			call.Arguments = strings.TrimSpace(lines[1])
		}
	}
	if call.Name == "" {
		return LLMToolCall{}, false
	}
	if call.Arguments == "" {
		call.Arguments = "{}"
	}
	return call, true
}

// parseXMLToolCalls extracts tool calls from <tool_call> blocks embedded in
// complete message content and returns them together with the content
// stripped of the blocks. Blocks without a usable name are ignored; if
// nothing parses the content is returned unchanged.
func parseXMLToolCalls(content string) ([]LLMToolCall, string) {
	matches := xmlToolCallRe.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return nil, content
	}

	calls := make([]LLMToolCall, 0, len(matches))
	for i, loc := range matches {
		if call, ok := parseXMLToolCallInner(content[loc[2]:loc[3]], i); ok {
			calls = append(calls, call)
		}
	}
	if len(calls) == 0 {
		return nil, content
	}
	return calls, xmlToolCallRe.ReplaceAllString(content, "")
}

// xmlToolCallGate keeps raw <tool_call> XML out of the text the user sees
// while it streams. Content is forwarded as soon as it cannot be the start
// of a block anymore; once an opener appears everything up to the closer is
// held back and, when the block completes, captured for adoption instead of
// being sent. The gate and xmlToolCallRe must recognize the same blocks.
type xmlToolCallGate struct {
	held   string
	blocks []string
}

// push feeds one stream chunk and returns the portion that is safe to
// forward to the user right away.
func (g *xmlToolCallGate) push(chunk string) string {
	g.held += chunk
	return g.release(false)
}

// finish flushes at end of stream: trailing non-block content is returned
// for forwarding, complete blocks are returned for adoption, and a block the
// model never closed is dropped (it never executes, so leaking its raw XML
// to the user would be pure noise).
func (g *xmlToolCallGate) finish() (string, []string) {
	return g.release(true), g.blocks
}

// release processes held content. final marks end of stream, where
// unterminated openers are dropped instead of waited on.
func (g *xmlToolCallGate) release(final bool) string {
	var forward strings.Builder
	for {
		if i := strings.Index(g.held, xmlToolCallOpen); i >= 0 {
			forward.WriteString(g.held[:i])
			if j := strings.Index(g.held[i:], xmlToolCallClose); j >= 0 {
				end := i + j + len(xmlToolCallClose)
				g.blocks = append(g.blocks, g.held[i:end])
				g.held = g.held[end:]
				continue
			}
			// opener present, closer not (yet): hold from the opener on
			if final {
				g.held = ""
			} else {
				g.held = g.held[i:]
			}
			return forward.String()
		}
		// no full opener: a partial opener at the tail could still become one
		if idx := partialXMLTagIndex(g.held, xmlToolCallOpen); idx >= 0 {
			forward.WriteString(g.held[:idx])
			g.held = g.held[idx:]
			if final {
				g.held = ""
			}
			return forward.String()
		}
		out := g.held
		g.held = ""
		return forward.String() + out
	}
}

// partialXMLTagIndex returns the earliest index where a proper prefix of tag
// begins, or -1. A proper prefix is shorter than the tag itself ("<tool_c"),
// so a full tag is not reported here.
func partialXMLTagIndex(s, tag string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != '<' {
			continue
		}
		rest := s[i:]
		if len(rest) >= len(tag) {
			rest = rest[:len(tag)]
		}
		if len(rest) < len(tag) && strings.HasPrefix(tag, rest) {
			return i
		}
	}
	return -1
}

// adoptInlineToolBlocks records captured <tool_call> blocks as pending tool
// calls (used when the provider sent no native tool_calls). The blocks were
// already kept out of the streamed text, so nothing needs stripping here.
func adoptInlineToolBlocks(l *LLM, tools *LLMToolState, blocks []string) bool {
	calls := make([]LLMToolCall, 0, len(blocks))
	for i, block := range blocks {
		inner := strings.TrimSuffix(strings.TrimPrefix(block, xmlToolCallOpen), xmlToolCallClose)
		if call, ok := parseXMLToolCallInner(inner, i); ok {
			calls = append(calls, call)
		}
	}
	if len(calls) == 0 {
		return false
	}
	logger.InfoCtx(l.Ctx, "inline tool_call blocks adopted", "count", len(calls))
	tools.Calls = append(tools.Calls, calls...)
	return true
}

func receiveLLMStream(ctx context.Context, l *LLM, roundStart time.Time, next func() (LLMStreamChunk, error), tools *LLMToolState) (*param.MsgInfo, bool, error) {
	provider, model := l.txtProviderType(), l.Model
	streamStart := time.Now()
	firstContent := true
	observeStream := func(err error) {
		metrics.LLMStreamDuration.WithLabelValues(provider, model, metrics.Outcome(ctx, err)).Observe(time.Since(streamStart).Seconds())
	}
	msgInfo := &param.MsgInfo{SendLen: FirstSendLen}
	hasTools := false
	var xmlGate xmlToolCallGate

	for {
		chunk, err := next()
		if errors.Is(err, io.EOF) {
			observeStream(nil)
			logger.InfoCtx(l.Ctx, "stream finished", "updateMsgID", l.MsgId)
			// flush what the gate still holds: trailing text goes out, the
			// captured <tool_call> blocks are adopted as tool calls
			if rest, blocks := xmlGate.finish(); len(blocks) > 0 || rest != "" {
				if rest != "" {
					msgInfo = l.SendMsg(msgInfo, rest)
				}
				if !hasTools {
					hasTools = adoptInlineToolBlocks(l, tools, blocks)
				}
			}
			if err := tools.execute(ctx, l); err != nil {
				return msgInfo, hasTools, err
			}
			return msgInfo, hasTools, nil
		}
		if err != nil {
			observeStream(err)
			logger.WarnCtx(l.Ctx, "stream error", "updateMsgID", l.MsgId, "err", err)
			return msgInfo, hasTools, err
		}

		for _, toolCall := range chunk.ToolCalls {
			hasTools = true
			if err := tools.observe(toolCall); err != nil {
				if errors.Is(err, ToolsJsonErr) {
					continue
				}
				observeStream(err)
				logger.ErrorCtx(l.Ctx, "request tools call error", "updateMsgID", l.MsgId, "err", err)
				return msgInfo, hasTools, err
			}
		}

		// stream content even in rounds that end in tool calls: the preamble
		// before a tool call is part of the answer the user should see (and
		// of WholeContent, which is replayed as the assistant message). The
		// final answer after the tool results streams as a fresh message from
		// the recursive Send. The gate withholds anything that may still turn
		// into a <tool_call> block so the raw XML never reaches the chat.
		if firstContent && chunk.Content != "" {
			metrics.LLMFirstContent.WithLabelValues(provider, model).Observe(time.Since(roundStart).Seconds())
			firstContent = false
		}
		if safe := xmlGate.push(chunk.Content); safe != "" {
			msgInfo = l.SendMsg(msgInfo, safe)
		}
		if chunk.HasUsage {
			l.Cs.Token += chunk.Usage
		}
	}
}

func sendLLMStreamOutput(l *LLM, msgInfo *param.MsgInfo, hasTools bool) {
	if l.MessageChan == nil {
		return
	}
	hasContent := msgInfo != nil && len(strings.TrimRightFunc(msgInfo.Content, unicode.IsSpace)) > 0
	if !hasContent && (!hasTools || !conf.BaseConfInfo.SendMcpRes) {
		return
	}
	if msgInfo == nil {
		msgInfo = &param.MsgInfo{}
	}
	if conf.BaseConfInfo.Powered != "" {
		msgInfo.Content += "\n\n" + conf.BaseConfInfo.Powered
	}
	l.MessageChan <- msgInfo
}
