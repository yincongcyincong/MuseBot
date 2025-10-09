package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtf16len(t *testing.T) {
	tests := map[string]int{
		"hello":   5,
		"你好":      2,
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

func TestDetectAudioFormat(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("ogg", DetectAudioFormat([]byte("OggS...........")))
	assert.Equal("mp3", DetectAudioFormat([]byte{0xFF, 0xFB, '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}))
	assert.Equal("wav", DetectAudioFormat(append([]byte("RIFF....WAVE....."), make([]byte, 4)...)))
	assert.Equal("unknown", DetectAudioFormat([]byte("??")))
}

func TestDetectImageFormat(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("jpeg", DetectImageFormat([]byte{0xFF, 0xD8, 0xFF, '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}))
	assert.Equal("png", DetectImageFormat([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}))
	assert.Equal("gif", DetectImageFormat([]byte("GIF87a..................")))
	assert.Equal("bmp", DetectImageFormat([]byte{0x42, 0x4D, '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.', '.'}))
	assert.Equal("unknown", DetectImageFormat([]byte("??")))
}

func TestValueToString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("123", ValueToString(123))
	assert.Equal("true", ValueToString(true))
	assert.Equal("hello", ValueToString("hello"))
	assert.Equal("1,2,3", ValueToString([]int{1, 2, 3}))
}

func TestMapKeysToString(t *testing.T) {
	assert := assert.New(t)
	m := map[string]int{"a": 1, "b": 2}
	result := MapKeysToString(m)
	assert.True(strings.Contains(result, "a"))
	assert.True(strings.Contains(result, "b"))
}
