package llm

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/param"
)

// a rejected action must surface as a real tool error (carrying the ❌ detail),
// not as success content: the task executor's step-failure accounting only
// counts errors, so swallowing them here would let the next plan step run
func TestExecuteTaskConfigToolSurfacesActionFailure(t *testing.T) {
	l := &LLM{ConfigExecutor: func(action *param.ConversationConfigAction,
		approve TaskApprovalFunc) (string, error) {
		return "", fmt.Errorf("configuration actions require an administrator user")
	}}

	content, err := l.executeTaskConfigTool(context.Background(), map[string]interface{}{
		"action": map[string]interface{}{"type": "skill", "operation": "list"},
	})
	assert.Error(t, err)
	assert.Empty(t, content)
	// the ❌ line stays in the error text: the tool loop feeds err back to the
	// model verbatim, so the failure reason is not lost
	assert.Contains(t, err.Error(), "❌ skill list")
	assert.Contains(t, err.Error(), "administrator user")
}

// one failing action among several marks the whole call failed, but the ✅
// results of the other actions ride along in the same message
func TestExecuteTaskConfigToolMixedActions(t *testing.T) {
	l := &LLM{ConfigExecutor: func(action *param.ConversationConfigAction,
		approve TaskApprovalFunc) (string, error) {
		if action.Operation == "delete" {
			return "", fmt.Errorf("not found")
		}
		return "done " + action.Operation, nil
	}}

	_, err := l.executeTaskConfigTool(context.Background(), map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{"type": "config", "operation": "get"},
			map[string]interface{}{"type": "skill", "operation": "delete"},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "✅ done get")
	assert.Contains(t, err.Error(), "❌ skill delete: not found")
}

func TestExecuteTaskConfigToolAllActionsSucceed(t *testing.T) {
	l := &LLM{ConfigExecutor: func(action *param.ConversationConfigAction,
		approve TaskApprovalFunc) (string, error) {
		return "ok", nil
	}}

	content, err := l.executeTaskConfigTool(context.Background(), map[string]interface{}{
		"action": map[string]interface{}{"type": "config", "operation": "get"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "✅ ok", content)
}
