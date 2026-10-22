package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func retryLLMRequest[T any](l *LLM, operation func() (T, error)) (T, error) {
	var ctx context.Context
	if l != nil {
		ctx = l.Ctx
	}
	return retryRequest(ctx, operation)
}

func observeLLMDuration(l *LLM, start time.Time) {
	metrics.APIRequestDuration.WithLabelValues(l.Model).Observe(time.Since(start).Seconds())
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

func (s *LLMToolState) observe(call LLMToolCall) error {
	if call.Name != "" || len(s.Calls) == 0 {
		s.Calls = append(s.Calls, LLMToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Type:      call.Type,
			Arguments: call.Arguments,
		})
	} else {
		latest := &s.Calls[len(s.Calls)-1]
		if call.ID != "" {
			latest.ID = call.ID
		}
		if call.Type != "" {
			latest.Type = call.Type
		}
		latest.Arguments += call.Arguments
	}

	latest := &s.Calls[len(s.Calls)-1]
	var property map[string]interface{}
	if err := json.Unmarshal([]byte(latest.Arguments), &property); err != nil {
		return ToolsJsonErr
	}
	return nil
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
			s.Results = append(s.Results, LLMToolResult{
				ID:      call.ID,
				Content: fmt.Sprintf("tool %s failed: invalid arguments: %v", call.Name, err),
			})
			continue
		}

		content, err := l.ExecMcpReq(ctx, call.Name, property)
		if err != nil {
			// a dead context means the request itself is being torn down, so
			// there is no point in feeding the failure back to the LLM
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			logger.ErrorCtx(ctx, "execute tool fail", "toolId", call.ID, "tool", call.Name, "err", err)
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

func receiveLLMStream(ctx context.Context, l *LLM, next func() (LLMStreamChunk, error), tools *LLMToolState) (*param.MsgInfo, bool, error) {
	msgInfo := &param.MsgInfo{SendLen: FirstSendLen}
	hasTools := false

	for {
		chunk, err := next()
		if errors.Is(err, io.EOF) {
			logger.InfoCtx(l.Ctx, "stream finished", "updateMsgID", l.MsgId)
			if err := tools.execute(ctx, l); err != nil {
				return msgInfo, hasTools, err
			}
			return msgInfo, hasTools, nil
		}
		if err != nil {
			logger.WarnCtx(l.Ctx, "stream error", "updateMsgID", l.MsgId, "err", err)
			return msgInfo, hasTools, err
		}

		for _, toolCall := range chunk.ToolCalls {
			hasTools = true
			if err := tools.observe(toolCall); err != nil {
				if errors.Is(err, ToolsJsonErr) {
					continue
				}
				logger.ErrorCtx(l.Ctx, "request tools call error", "updateMsgID", l.MsgId, "err", err)
				return msgInfo, hasTools, err
			}
		}

		if !hasTools {
			msgInfo = l.SendMsg(msgInfo, chunk.Content)
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
