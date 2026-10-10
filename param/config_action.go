package param

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type ConversationConfigID int64

func (id *ConversationConfigID) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if string(trimmed) == "null" {
		*id = 0
		return nil
	}

	if bytes.HasPrefix(trimmed, []byte(`"`)) {
		var text string
		if err := json.Unmarshal(trimmed, &text); err != nil {
			return err
		}
		value, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		if err != nil {
			return err
		}
		*id = ConversationConfigID(value)
		return nil
	}

	var value int64
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return err
	}
	*id = ConversationConfigID(value)
	return nil
}

func (id ConversationConfigID) Int64() int64 {
	return int64(id)
}

type ConversationConfigAction struct {
	Type         string               `json:"type"`
	ConfigType   string               `json:"config_type,omitempty"`
	Key          string               `json:"key,omitempty"`
	Value        json.RawMessage      `json:"value,omitempty"`
	Operation    string               `json:"operation,omitempty"`
	ID           ConversationConfigID `json:"id,omitempty"`
	Name         string               `json:"name,omitempty"`
	Path         string               `json:"path,omitempty"`
	Description  string               `json:"description,omitempty"`
	Instructions string               `json:"instructions,omitempty"`
	Cron         string               `json:"cron,omitempty"`
	Prompt       string               `json:"prompt,omitempty"`
	Command      string               `json:"command,omitempty"`
	Server       json.RawMessage      `json:"server,omitempty"`
	Args         []string             `json:"args,omitempty"`
	Timeout      int                  `json:"timeout,omitempty"`
	Content      string               `json:"content,omitempty"`
	OldText      string               `json:"old_text,omitempty"`
	NewText      string               `json:"new_text,omitempty"`
	Target       string               `json:"target,omitempty"`
}

type ConversationConfigPlan struct {
	Actions []*ConversationConfigAction `json:"actions"`
}

type ConversationConfigIntent struct {
	Config bool `json:"config"`
}

func ParseConversationConfigIntent(content string) (*ConversationConfigIntent, error) {
	start := strings.Index(content, "{")
	if start < 0 {
		return nil, fmt.Errorf("configuration intent does not contain JSON")
	}

	decoder := json.NewDecoder(strings.NewReader(content[start:]))
	intent := new(ConversationConfigIntent)
	if err := decoder.Decode(intent); err != nil {
		return nil, fmt.Errorf("decode configuration intent: %w", err)
	}
	return intent, nil
}

func ParseConversationConfigPlan(content string) (*ConversationConfigPlan, error) {
	start := strings.Index(content, "{")
	if start < 0 {
		return nil, fmt.Errorf("configuration plan does not contain JSON")
	}

	decoder := json.NewDecoder(strings.NewReader(content[start:]))
	plan := new(ConversationConfigPlan)
	if err := decoder.Decode(plan); err != nil {
		return nil, fmt.Errorf("decode configuration plan: %w", err)
	}
	if len(plan.Actions) == 0 {
		return nil, fmt.Errorf("configuration plan has no actions")
	}
	if len(plan.Actions) > 20 {
		return nil, fmt.Errorf("configuration plan has too many actions")
	}

	var invalidTypes []string
	for _, action := range plan.Actions {
		if action == nil || action.Type == "" {
			invalidTypes = append(invalidTypes, "empty")
			continue
		}
		switch action.Type {
		case "config", "mcp", "skill", "cron", "command", "file":
		default:
			invalidTypes = append(invalidTypes, action.Type)
		}
	}
	if len(invalidTypes) > 0 {
		return nil, fmt.Errorf("unsupported configuration action type: %s", strings.Join(invalidTypes, ", "))
	}

	return plan, nil
}

func (a *ConversationConfigAction) HasValue() bool {
	return len(bytes.TrimSpace(a.Value)) != 0
}
