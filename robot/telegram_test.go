package robot

import (
	"strings"
	"testing"
	
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
