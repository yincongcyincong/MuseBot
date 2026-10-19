package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
)

// executeTaskConfigTool is the built-in bot_config tool of the /task workflow.
// The model passes one action object (or an actions array); each action is
// validated with the shared plan parser and executed through the injected
// ConfigExecutor, which enforces the admin gate and the safety boundaries.
func (l *LLM) executeTaskConfigTool(ctx context.Context, property map[string]interface{}) (string, error) {
	if l.ConfigExecutor == nil {
		return "", fmt.Errorf("config executor is unavailable in this context")
	}

	rawActions, ok := configToolActions(property)
	if !ok {
		return "", fmt.Errorf("bot_config requires an action object or an actions array")
	}

	plan, err := param.ParseConversationConfigPlan(string(rawActions))
	if err != nil {
		return "", fmt.Errorf("invalid bot_config action: %w", err)
	}

	results := make([]string, 0, len(plan.Actions))
	for _, action := range plan.Actions {
		logger.InfoCtx(ctx, "task config action", "type", action.Type, "operation", action.Operation,
			"key", action.Key, "path", action.Path)
		result, err := l.ConfigExecutor(action)
		if err != nil {
			logger.ErrorCtx(ctx, "task config action fail", "type", action.Type,
				"operation", action.Operation, "err", err)
			results = append(results, fmt.Sprintf("❌ %s %s: %s", action.Type, action.Operation, err.Error()))
			continue
		}
		results = append(results, "✅ "+result)
	}

	return strings.Join(results, "\n"), nil
}

// configToolActions normalizes the tool arguments into a plan JSON document.
func configToolActions(property map[string]interface{}) (json.RawMessage, bool) {
	if actions, ok := property["actions"].([]interface{}); ok {
		data, err := json.Marshal(map[string]interface{}{"actions": actions})
		if err != nil {
			return nil, false
		}
		return data, true
	}
	if action, ok := property["action"].(map[string]interface{}); ok {
		data, err := json.Marshal(map[string]interface{}{"actions": []interface{}{action}})
		if err != nil {
			return nil, false
		}
		return data, true
	}
	// tolerate the model inlining the action fields directly on the tool call
	if t, _ := property["type"].(string); t != "" {
		if _, operation := property["operation"].(string); operation {
			data, err := json.Marshal(map[string]interface{}{"actions": []interface{}{property}})
			if err != nil {
				return nil, false
			}
			return data, true
		}
	}
	return nil, false
}
