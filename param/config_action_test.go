package param

import "testing"

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

func TestParseConversationConfigIntent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "true", content: "prefix {\"config\":true} suffix", want: true},
		{name: "false", content: "{\"config\":false}", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent, err := ParseConversationConfigIntent(tt.content)
			if err != nil {
				t.Fatal(err)
			}
			if intent.Config != tt.want {
				t.Errorf("config = %v, want %v", intent.Config, tt.want)
			}
		})
	}
}
