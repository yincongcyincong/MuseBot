package robot

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
	tel := NewTelegramRobot(updatePrivate, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	})
	if skip := tel.skipThisMsg(); skip {
		t.Error("private chat message should not be skipped")
	}
	
	updateGroup := makeFakeUpdateWithText("hello", "", "group")
	tel = NewTelegramRobot(updateGroup, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	})
	if skip := tel.skipThisMsg(); !skip {
		t.Error("group message without mention should be skipped")
	}
	
	updateGroupMention := makeFakeUpdateWithText("hello @"+fakeBotUserName, fakeBotUserName, "group")
	tel = NewTelegramRobot(updateGroupMention, &tgbotapi.BotAPI{
		Self: tgbotapi.User{UserName: fakeBotUserName},
	})
	if skip := tel.skipThisMsg(); skip {
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
