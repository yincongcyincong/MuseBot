package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func TestExecuteTaskConfigToolDispatchesActions(t *testing.T) {
	executed := make([]*param.ConversationConfigAction, 0)
	l := &LLM{
		ConfigExecutor: func(action *param.ConversationConfigAction) (string, error) {
			executed = append(executed, action)
			return action.Operation + " done", nil
		},
	}

	// a single action object
	out, err := l.executeTaskConfigTool(context.Background(), map[string]interface{}{
		"action": map[string]interface{}{"type": "config", "operation": "list"},
	})
	assert.NoError(t, err)
	assert.Contains(t, out, "✅ list done")
	assert.Len(t, executed, 1)

	// an actions array, including the "op" alias
	executed = executed[:0]
	out, err = l.executeTaskConfigTool(context.Background(), map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{"type": "file", "op": "list", "path": "doc"},
			map[string]interface{}{"type": "cron", "operation": "list"},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, executed, 2)
	assert.Equal(t, "list", executed[0].Operation) // alias resolved
	assert.Equal(t, "doc", executed[0].Path)
	assert.Equal(t, 2, strings.Count(out, "✅"))

	// fields inlined directly on the tool call
	executed = executed[:0]
	out, err = l.executeTaskConfigTool(context.Background(), map[string]interface{}{
		"type": "mcp", "operation": "sync",
	})
	assert.NoError(t, err)
	assert.Len(t, executed, 1)
	assert.Equal(t, "mcp", executed[0].Type)

	// unsupported action types are rejected before reaching the executor
	_, err = l.executeTaskConfigTool(context.Background(), map[string]interface{}{
		"action": map[string]interface{}{"type": "reboot", "operation": "now"},
	})
	assert.ErrorContains(t, err, "unsupported configuration action type")

	// no executor injected -> explicit error instead of falling through to MCP
	bare := &LLM{}
	_, err = bare.ExecMcpReq(context.Background(), conf.BuiltinConfigTool,
		map[string]interface{}{"action": map[string]interface{}{"type": "config", "operation": "list"}})
	assert.ErrorContains(t, err, "config executor is unavailable")
}
