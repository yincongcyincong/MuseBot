package robot

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 为了测试 handleCommand 逻辑，我们需要构造一个模拟的 BotAPI 实例。
// 这里我们简单地构造一个 fakeBot 实现 Send 方法，记录调用信息。
type fakeBot struct {
	sentMessages []tgbotapi.Chattable
}

func (f *fakeBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	f.sentMessages = append(f.sentMessages, c)
	// 简单返回一个 Message，其 MessageID 为1，方便测试。
	return tgbotapi.Message{MessageID: 1}, nil
}

// 模拟 Request 方法（在我们的代码中也被使用）
func (f *fakeBot) Request(c tgbotapi.Chattable) (tgbotapi.APIResponse, error) {
	return tgbotapi.APIResponse{Ok: true}, nil
}

// 为方便测试，我们构造一个假的 Update 对象
func makeFakeUpdateWithCommand(command string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "/" + command,
			Chat: &tgbotapi.Chat{
				ID:   12345,
				Type: "private",
			},
		},
	}
}

func makeFakeUpdateWithText(text string, botUserName string, chatType string) tgbotapi.Update {
	// 如果是群组，需要包含@botUserName，否则 skipThisMsg 会跳过消息
	if chatType != "private" && !strings.Contains(text, "@"+botUserName) {
		text = text + " @" + botUserName
	}
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: text,
			Chat: &tgbotapi.Chat{
				ID:   12345,
				Type: chatType,
			},
		},
	}
}

// 单测 skipThisMsg 函数
func TestSkipThisMsg(t *testing.T) {
	// 构造 fake bot 自定义 botSelf.UserName
	fakeBotUserName := "TestBot"
	// 构造 update 对象
	// 1. 私聊不跳过
	updatePrivate := makeFakeUpdateWithText("hello", fakeBotUserName, "private")
	if skip := skipThisMsg(updatePrivate, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	}); skip {
		t.Error("private chat message should not be skipped")
	}
	// 2. 群组消息且不包含 @botUserName 跳过
	updateGroup := makeFakeUpdateWithText("hello", "", "group")
	if skip := skipThisMsg(updateGroup, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	}); !skip {
		t.Error("group message without mention should be skipped")
	}
	// 3. 群组消息包含 @botUserName 不跳过
	updateGroupMention := makeFakeUpdateWithText("hello @"+fakeBotUserName, fakeBotUserName, "group")
	if skip := skipThisMsg(updateGroupMention, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	}); skip {
		t.Error("group message with mention should not be skipped")
	}
}

// 单测 sleepUtilNoLimit 函数
func TestSleepUtilNoLimit(t *testing.T) {
	// 构造一个 Telegram API error 模拟 Too Many Requests 情况
	apiErr := tgbotapi.Error{
		Message: "Too Many Requests",
		ResponseParameters: tgbotapi.ResponseParameters{
			RetryAfter: 1,
		}, // 1 秒
	}
	// 包装错误，使得 errors.As 能够匹配到 *tgbotapi.Error
	wrappedErr := fmt.Errorf("wrapped: %w", &apiErr)

	start := time.Now()
	if !sleepUtilNoLimit(1, wrappedErr) {
		t.Error("Expected sleepUtilNoLimit to return true on rate limit error")
	}
	elapsed := time.Since(start)
	if elapsed < time.Second {
		t.Errorf("Expected sleep duration at least 1 second, got %v", elapsed)
	}

	// 测试非 rate limit 错误
	otherErr := errors.New("some other error")
	if sleepUtilNoLimit(1, otherErr) {
		t.Error("Expected sleepUtilNoLimit to return false on non rate limit error")
	}
}
