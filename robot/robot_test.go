package robot

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type fakeBot struct {
	sentMessages []tgbotapi.Chattable
}

func (f *fakeBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	f.sentMessages = append(f.sentMessages, c)
	// 简单返回一个 Message，其 MessageID 为1，方便测试。
	return tgbotapi.Message{MessageID: 1}, nil
}

func (f *fakeBot) Request(c tgbotapi.Chattable) (tgbotapi.APIResponse, error) {
	return tgbotapi.APIResponse{Ok: true}, nil
}

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

func TestSkipThisMsg(t *testing.T) {
	fakeBotUserName := "TestBot"

	updatePrivate := makeFakeUpdateWithText("hello", fakeBotUserName, "private")
	if skip := skipThisMsg(updatePrivate, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	}); skip {
		t.Error("private chat message should not be skipped")
	}

	updateGroup := makeFakeUpdateWithText("hello", "", "group")
	if skip := skipThisMsg(updateGroup, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	}); !skip {
		t.Error("group message without mention should be skipped")
	}

	updateGroupMention := makeFakeUpdateWithText("hello @"+fakeBotUserName, fakeBotUserName, "group")
	if skip := skipThisMsg(updateGroupMention, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	}); skip {
		t.Error("group message with mention should not be skipped")
	}
}

func TestSleepUtilNoLimit(t *testing.T) {
	apiErr := tgbotapi.Error{
		Message: "Too Many Requests",
		ResponseParameters: tgbotapi.ResponseParameters{
			RetryAfter: 1,
		},
	}
	wrappedErr := fmt.Errorf("wrapped: %w", &apiErr)

	start := time.Now()
	if !sleepUtilNoLimit(1, wrappedErr) {
		t.Error("Expected sleepUtilNoLimit to return true on rate limit error")
	}
	elapsed := time.Since(start)
	if elapsed < time.Second {
		t.Errorf("Expected sleep duration at least 1 second, got %v", elapsed)
	}

	otherErr := errors.New("some other error")
	if sleepUtilNoLimit(1, otherErr) {
		t.Error("Expected sleepUtilNoLimit to return false on non rate limit error")
	}
}
