package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yincongcyincong/MuseBot/i18n"
)

const taskToolApprovalTimeout = 5 * time.Minute

type taskToolApproval struct {
	Token     string
	Operation string
	Path      string
	Detail    string
	response  chan bool
}

var taskToolApprovals sync.Map

func taskApprovalOwner(userID, chatID string) string {
	return userID + "\x00" + chatID
}

func requestTaskToolApproval(ctx context.Context, l *LLM, operation, target, detail string) (bool, error) {
	approval := &taskToolApproval{
		Token:     strings.ReplaceAll(uuid.NewString(), "-", "")[:8],
		Operation: operation,
		Path:      target,
		Detail:    detail,
		response:  make(chan bool, 1),
	}

	owner := taskApprovalOwner(l.UserId, l.ChatId)
	if existing, loaded := taskToolApprovals.LoadOrStore(owner, approval); loaded {
		pending := existing.(*taskToolApproval)
		return false, fmt.Errorf("another tool approval %s is pending", pending.Token)
	}

	l.DirectSendMsg(i18n.GetMessage("task_tool_approval_request", map[string]interface{}{
		"operation": approval.Operation,
		"path":      approval.Path,
		"detail":    approval.Detail,
		"timeout":   int(taskToolApprovalTimeout.Minutes()),
	}), true)

	defer func() {
		if current, loaded := taskToolApprovals.Load(owner); loaded && current.(*taskToolApproval) == approval {
			taskToolApprovals.Delete(owner)
		}
	}()

	timeout := time.NewTimer(taskToolApprovalTimeout)
	defer timeout.Stop()

	select {
	case approved := <-approval.response:
		return approved, nil
	case <-timeout.C:
		return false, fmt.Errorf("file tool approval expired")
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func ResolveTaskToolApproval(userID, chatID, token string, approved bool) bool {
	owner := taskApprovalOwner(userID, chatID)
	value, loaded := taskToolApprovals.Load(owner)
	if !loaded {
		return false
	}

	approval := value.(*taskToolApproval)
	if token != "" && approval.Token != token {
		return false
	}
	taskToolApprovals.Delete(owner)
	approval.response <- approved
	return true
}
