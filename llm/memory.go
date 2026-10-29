package llm

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
)

// errNoMemoryJSON covers both "model replied without a JSON array" and an
// empty result: either way the stored memories stay as they were.
var errNoMemoryJSON = errors.New("memory merge returned no JSON array")

// memoryExtractBudget caps how many recent QA pairs feed the consolidation
// prompt: enough context to catch durable facts, small enough to stay cheap.
const memoryExtractBudget = 8

// lastMemoryExtract throttles consolidation per user (userId -> time.Time).
var lastMemoryExtract sync.Map

var jsonArrayRe = regexp.MustCompile(`\[[\s\S]*\]`)

// MaybeConsolidateMemories merges the user's long-term memories with their
// recent conversation into a new memory set, at most once per
// MemoryInterval per user. Any failure leaves the stored memories untouched.
func MaybeConsolidateMemories(ctx context.Context, userId string) {
	if !conf.BaseConfInfo.EnableMemory || userId == "" {
		return
	}

	interval := time.Duration(conf.BaseConfInfo.MemoryInterval) * time.Second
	if interval <= 0 {
		interval = time.Hour
	}
	if last, ok := lastMemoryExtract.Load(userId); ok {
		if t, ok := last.(time.Time); ok && time.Since(t) < interval {
			return
		}
	}
	lastMemoryExtract.Store(userId, time.Now())

	msgRecords := db.GetMsgRecord(userId)
	if msgRecords == nil || len(msgRecords.AQs) == 0 {
		return
	}

	start := len(msgRecords.AQs) - memoryExtractBudget
	if start < 0 {
		start = 0
	}
	if err := consolidateMemories(ctx, userId, msgRecords.AQs[start:]); err != nil {
		// a trivial conversation ("hi") or a chatty model often answers in
		// prose instead of JSON; stored memories stay untouched, so this is
		// a normal outcome rather than a fault
		logger.InfoCtx(ctx, "consolidate memories skip, keep old memories", "err", err)
	}
}

// consolidateMemories runs the merge prompt and replaces the stored set.
func consolidateMemories(ctx context.Context, userId string, recentAQs []*db.AQ) error {
	var memBuilder strings.Builder
	for _, m := range db.GetMemories(userId) {
		memBuilder.WriteString("- " + m.Content + "\n")
	}
	var histBuilder strings.Builder
	for _, aq := range recentAQs {
		if aq.Question != "" {
			histBuilder.WriteString("Q: " + aq.Question + "\n")
		}
		if aq.Answer != "" {
			histBuilder.WriteString("A: " + aq.Answer + "\n")
		}
	}

	maxCount := conf.BaseConfInfo.MemoryMaxCount
	if maxCount <= 0 {
		maxCount = 20
	}
	prompt := i18n.GetMessage("memory_merge_prompt", map[string]interface{}{
		"memories":  memBuilder.String(),
		"history":   histBuilder.String(),
		"max_count": maxCount,
	})

	l := NewLLM(WithUserId(userId), WithContent(prompt), WithContext(ctx), WithUtilityModel())
	l.Cs.UseRecord = false
	l.LLMClient.GetModel(l)
	l.LLMClient.GetMessage(openai.ChatMessageRoleUser, prompt)

	metrics.APIRequestCount.WithLabelValues(l.Model).Inc()
	content, err := l.LLMClient.SyncSend(ctx, l)
	if err != nil {
		return err
	}

	match := jsonArrayRe.FindString(content)
	if match == "" {
		return errNoMemoryJSON
	}
	var items []string
	if err = json.Unmarshal([]byte(match), &items); err != nil {
		return err
	}

	trimmed := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			trimmed = append(trimmed, item)
		}
	}
	if len(trimmed) > maxCount {
		trimmed = trimmed[:maxCount]
	}
	if len(trimmed) == 0 {
		return errNoMemoryJSON
	}

	if err = db.ReplaceMemories(ctx, userId, trimmed); err != nil {
		return err
	}
	logger.InfoCtx(ctx, "consolidate memories done", "userID", userId, "count", len(trimmed))
	return nil
}
