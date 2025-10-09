package utils

import (
	"testing"
)

func TestDecreaseUserChat(t *testing.T) {
	userId := "999999999"
	// 初始化次数为 3
	userChatMap.Store(userId, 3)

	DecreaseUserChat(userId)

	if val, ok := userChatMap.Load(userId); !ok || val.(int) != 2 {
		t.Errorf("Expected times to be 2, got %v", val)
	}
}
