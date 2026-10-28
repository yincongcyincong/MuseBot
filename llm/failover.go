package llm

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

// failoverTarget is one rung of the failover ladder.
type failoverTarget struct {
	Type  string
	Model string
}

// sendWithFailover runs the primary provider with the full retry budget
// (the historical behavior), then walks the failover chain with one
// attempt per rung. A nil return means the answer was delivered.
func (l *LLM) sendWithFailover() error {
	retryTimes := conf.BaseConfInfo.LLMRetryTimes
	if retryTimes <= 0 {
		retryTimes = 1
	}
	var err error
	for attempt := 0; attempt < retryTimes; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(conf.BaseConfInfo.LLMRetryInterval) * time.Millisecond)
		}
		err = l.attempt()
		if err == nil {
			return nil
		}
		// a cancelled or expired context fails every attempt instantly, so
		// retrying cannot succeed; surface the error right away
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// auth/quota/rate-limit errors cannot be retried into success on
		// the same provider; jump straight to the fallback chain
		if l.canFailover() && shouldFailoverNow(err) {
			break
		}
		logger.WarnCtx(l.Ctx, "llm request fail, retrying", "attempt", attempt+1, "err", err)
	}

	if err == nil || !l.canFailover() {
		return err
	}

	primary := l.txtProviderType()
	for _, target := range l.failoverChain() {
		if ctxErr := l.Ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if !l.canFailover() {
			break
		}
		l.providerOverride = target.Type
		l.modelOverride = target.Model
		l.LLMClient.GetModel(l)
		logger.WarnCtx(l.Ctx, "failover to backup provider",
			"from", primary, "to", target.Type, "model", target.Model)
		metrics.APIFailoverCount.WithLabelValues(primary, target.Type).Inc()

		err = l.attempt()
		if err == nil {
			logger.InfoCtx(l.Ctx, "failover success", "from", primary, "to", target.Type)
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}
	return err
}

// attempt runs one LLM round, streaming or sync per configuration. On the
// sync path the full answer is pushed to the message channel, matching the
// historical CallLLM behavior.
func (l *LLM) attempt() error {
	if conf.BaseConfInfo.IsStreaming {
		return l.LLMClient.Send(l.Ctx, l)
	}
	content, err := l.LLMClient.SyncSend(l.Ctx, l)
	if err == nil {
		l.MessageChan <- &param.MsgInfo{Content: content}
		l.WholeContent = content
	}
	return err
}

// canFailover reports whether the current turn may still switch providers:
// failover enabled, an OpenAI-compatible client (whose message state
// carries over), no pinned override (utility requests keep their model),
// and no streamed output yet (a restart would duplicate what the user
// already received).
func (l *LLM) canFailover() bool {
	if !conf.BaseConfInfo.EnableFailover || l.providerOverride != "" {
		return false
	}
	if _, ok := l.LLMClient.(*OpenAIReq); !ok {
		return false
	}
	if conf.BaseConfInfo.IsStreaming && l.WholeContent != "" {
		return false
	}
	return true
}

// txtProviderType is the provider type this chat request resolves to.
func (l *LLM) txtProviderType() string {
	if l.providerOverride != "" {
		return l.providerOverride
	}
	return utils.GetTxtType(db.GetCtxUserInfo(l.Ctx).LLMConfigRaw)
}

// failoverChain lists the backup providers after the primary: the explicit
// FailoverProviders order when set, else every other configured
// OpenAI-compatible type. Ollama is always skipped: its protocol client
// keeps messages in its own format, so mid-turn migration is not possible.
func (l *LLM) failoverChain() []failoverTarget {
	primary := l.txtProviderType()

	var types []string
	if explicit := conf.BaseConfInfo.FailoverProviders; explicit != "" {
		for _, t := range strings.Split(explicit, ",") {
			t = strings.TrimSpace(t)
			if t != "" && t != primary && t != param.Ollama {
				types = append(types, t)
			}
		}
	} else {
		for _, t := range utils.GetAvailTxtType() {
			if t != primary && t != param.Ollama {
				types = append(types, t)
			}
		}
	}

	targets := make([]failoverTarget, 0, len(types))
	for _, t := range types {
		if _, ok := openaiEndpoints[t]; !ok {
			continue // no OpenAI-compatible endpoint for this type
		}
		targets = append(targets, failoverTarget{Type: t, Model: utils.GetTxtModel(t)})
	}
	return targets
}

// shouldFailoverNow reports errors where more retries on the same provider
// cannot succeed (auth/quota) or are not worth waiting for (rate limit).
// go-openai returns *APIError for JSON error bodies and *RequestError
// otherwise (e.g. a proxy replying without one); both carry the status.
func shouldFailoverNow(err error) bool {
	statusCode := 0
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		statusCode = apiErr.HTTPStatusCode
	}
	var reqErr *openai.RequestError
	if errors.As(err, &reqErr) {
		statusCode = reqErr.HTTPStatusCode
	}

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests:
		return true
	}
	return false
}
