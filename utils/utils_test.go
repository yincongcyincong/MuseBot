package utils

import (
	"testing"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

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

func TestMD5(t *testing.T) {
	input := "hello world"
	want := "5eb63bbbe01eeed093cb22bb8f5acdc3"
	got := MD5(input)
	if got != want {
		t.Errorf("MD5(%q) = %s; want %s", input, got, want)
	}
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
