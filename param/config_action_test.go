package param

import (
	"errors"
	"testing"
)

func TestParseConversationConfigPlan(t *testing.T) {
	content := "```json\n{\"actions\":[{\"type\":\"config\",\"operation\":\"set\",\"config_type\":\"base\",\"key\":\"smart_mode\",\"value\":true}]}\n```"
	plan, err := ParseConversationConfigPlan(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 1 {
		t.Fatalf("actions = %d, want 1", len(plan.Actions))
	}
	action := plan.Actions[0]
	if action.Type != "config" || action.Operation != "set" || !action.HasValue() {
		t.Errorf("unexpected action: %+v", action)
	}
}

func TestParseConversationConfigPlanRejectsUnsupportedType(t *testing.T) {
	if _, err := ParseConversationConfigPlan(`{"actions":[{"type":"shell"}]}`); err == nil {
		t.Fatal("expected unsupported action type error")
	}
}

func TestParseConversationConfigPlanBuiltInActions(t *testing.T) {
	content := `{"actions":[
		{"type":"command","operation":"run","args":["git","status"],"timeout":30},
		{"type":"file","operation":"edit","path":"doc.md","old_text":"old","new_text":"new"}
	]}`
	plan, err := ParseConversationConfigPlan(content)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].Type != "command" || plan.Actions[0].Args[0] != "git" || plan.Actions[0].Timeout != 30 {
		t.Errorf("unexpected command action: %+v", plan.Actions[0])
	}
	if plan.Actions[1].Type != "file" || plan.Actions[1].OldText != "old" || plan.Actions[1].NewText != "new" {
		t.Errorf("unexpected file action: %+v", plan.Actions[1])
	}
}

func TestParseConversationConfigPlanAcceptsQuotedCronID(t *testing.T) {
	content := `{"actions":[{"type":"cron","operation":"delete","id":"42"}]}`
	plan, err := ParseConversationConfigPlan(content)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].ID.Int64() != 42 {
		t.Fatalf("id = %d, want 42", plan.Actions[0].ID.Int64())
	}
}

// an empty plan must be distinguishable from real failures so callers can
// treat it like {"config":false} and fall back to normal chat
func TestParseConversationConfigPlanEmptyActions(t *testing.T) {
	for _, content := range []string{`{"actions":[]}`, `{}`, `{"actions":null}`} {
		_, err := ParseConversationConfigPlan(content)
		if !errors.Is(err, ErrEmptyConfigPlan) {
			t.Errorf("ParseConversationConfigPlan(%s) err = %v, want ErrEmptyConfigPlan", content, err)
		}
	}
}

// models sometimes emit "op" instead of "operation"; accept it as an alias
func TestParseConversationConfigPlanAcceptsOpAlias(t *testing.T) {
	plan, err := ParseConversationConfigPlan(`{"actions":[{"type":"file","op":"list","path":"doc"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Actions) != 1 || plan.Actions[0].Operation != "list" {
		t.Fatalf("operation = %q, want list", plan.Actions[0].Operation)
	}

	// explicit "operation" still wins over the alias
	plan, err = ParseConversationConfigPlan(`{"actions":[{"type":"config","op":"get","operation":"list"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Actions[0].Operation != "list" {
		t.Fatalf("operation = %q, want list (explicit key must win)", plan.Actions[0].Operation)
	}
}
