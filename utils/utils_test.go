package utils

import (
	"fmt"
	"net/http"
	"testing"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/test"
)

func TestGetChatIdAndMsgIdAndUserName_MessageUpdate(t *testing.T) {
	chatID := int64(123)
	msgID := 456
	userID := int64(789)
	
	updateMsg := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: chatID},
			MessageID: msgID,
			From:      &tgbotapi.User{ID: userID},
		},
	}
	cid, mid, uid := GetChatIdAndMsgIdAndUserID(updateMsg)
	if cid != chatID || mid != msgID || uid != userID {
		t.Errorf("Expected (%d,%d,%d), got (%d,%d,%d)", chatID, msgID, userID, cid, mid, uid)
	}
	
	updateCallback := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat:      &tgbotapi.Chat{ID: chatID},
				MessageID: msgID,
			},
			From: &tgbotapi.User{ID: userID},
		},
	}
	cid, mid, uid = GetChatIdAndMsgIdAndUserID(updateCallback)
	if cid != chatID || mid != msgID || uid != userID {
		t.Errorf("Expected (%d,%d,%d), got (%d,%d,%d)", chatID, msgID, userID, cid, mid, uid)
	}
}

func TestUtf16len(t *testing.T) {
	tests := map[string]int{
		"hello":   5,
		"你好":    2,
		"𠀀":       2, // surrogate pair in utf16
		"":        0,
		"abc𠀀def": 8,
	}
	for input, expected := range tests {
		if got := Utf16len(input); got != expected {
			t.Errorf("Utf16len(%q) = %d; want %d", input, got, expected)
		}
	}
}

func TestParseInt(t *testing.T) {
	if got := ParseInt("123"); got != 123 {
		t.Errorf("ParseInt(\"123\") = %d; want 123", got)
	}
	if got := ParseInt("abc"); got != 0 {
		t.Errorf("ParseInt(\"abc\") = %d; want 0", got)
	}
}

func TestReplaceCommand(t *testing.T) {
	content := "/start @botname some text"
	command := "/start"
	botName := "botname"
	want := "some text"
	got := ReplaceCommand(content, command, botName)
	if got != want {
		t.Errorf("ReplaceCommand() = %q; want %q", got, want)
	}
}

func TestGetChatIdAndMsgIdAndUserName_CallbackQueryUpdate(t *testing.T) {
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			Message: &tgbotapi.Message{
				Chat:      &tgbotapi.Chat{ID: 654321},
				MessageID: 456,
			},
			From: &tgbotapi.User{ID: 111},
		},
	}
	
	chatId, msgId, userId := GetChatIdAndMsgIdAndUserID(update)
	assert.Equal(t, int64(654321), chatId)
	assert.Equal(t, 456, msgId)
	assert.Equal(t, int64(111), userId)
}

func TestMD5(t *testing.T) {
	input := "hello world"
	want := "5eb63bbbe01eeed093cb22bb8f5acdc3"
	got := MD5(input)
	if got != want {
		t.Errorf("MD5(%q) = %s; want %s", input, got, want)
	}
}

func TestGetChatIdAndMsgIdAndUserName_EmptyUpdate(t *testing.T) {
	update := tgbotapi.Update{}
	
	chatId, msgId, userId := GetChatIdAndMsgIdAndUserID(update)
	assert.Equal(t, int64(0), chatId)
	assert.Equal(t, 0, msgId)
	assert.Equal(t, int64(0), userId)
}

func TestCheckMsgIsCallback(t *testing.T) {
	updateWithCallback := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{},
	}
	assert.True(t, CheckMsgIsCallback(updateWithCallback))
	
	updateWithMessage := tgbotapi.Update{
		Message: &tgbotapi.Message{},
	}
	assert.False(t, CheckMsgIsCallback(updateWithMessage))
	
	updateEmpty := tgbotapi.Update{}
	assert.False(t, CheckMsgIsCallback(updateEmpty))
}

func TestGetPhotoContent(t *testing.T) {
	
	// 调用被测函数
	mockClient := &test.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// 返回模拟响应
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: test.NewMockBody(`{
  "ok": true,
  "result": {
    "file_id": "ABCD1234",
    "file_unique_id": "XYZ9876",
    "file_size": 1024,
	"file_path": "a/b/c.jpg"
  }
}`),
			}, nil
		},
	}
	bot := &tgbotapi.BotAPI{
		Token:  "xxxx",
		Client: mockClient,
	}
	bot.SetAPIEndpoint(tgbotapi.APIEndpoint)
	
	photos := []tgbotapi.PhotoSize{
		{FileID: "small", FileSize: 1024},
		{FileID: "large", FileSize: 7 * 1024 * 1024},
	}
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Photo: photos,
		},
	}
	
	tt := ""
	conf.BaseConfInfo.TelegramProxy = &tt
	
	byteContent := GetPhotoContent(update, bot)
	
	fmt.Println(byteContent)
	
}
