package rag

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDocumentText(t *testing.T) {
	text, err := ParseDocument("notes.txt", strings.NewReader("hello document"))
	assert.NoError(t, err)
	assert.Contains(t, text, "hello document")
}

func TestParseDocumentUnknownExtFallsBackToText(t *testing.T) {
	text, err := ParseDocument("deploy.bat", strings.NewReader("@echo off\r\necho hello\r\n"))
	assert.NoError(t, err)
	assert.Contains(t, text, "echo hello")

	text, err = ParseDocument("changes.diff", strings.NewReader("--- a/x\n+++ b/x\n+new line\n"))
	assert.NoError(t, err)
	assert.Contains(t, text, "new line")

	// no extension at all
	text, err = ParseDocument("Makefile", strings.NewReader("all: build\n"))
	assert.NoError(t, err)
	assert.Contains(t, text, "all: build")
}

func TestParseDocumentBinaryRejected(t *testing.T) {
	_, err := ParseDocument("archive.zip", strings.NewReader("PK\x03\x04\x00\x00\x00\x01\x00"))
	assert.Error(t, err)
}

func TestParseDocumentCSV(t *testing.T) {
	text, err := ParseDocument("data.csv", strings.NewReader("name,age\nalice,30\n"))
	assert.NoError(t, err)
	assert.Contains(t, text, "alice")
}

func TestParseDocumentHTML(t *testing.T) {
	text, err := ParseDocument("page.html", strings.NewReader("<html><body><h1>Title</h1><p>Body text</p></body></html>"))
	assert.NoError(t, err)
	assert.Contains(t, text, "Title")
}

func TestTruncateDocText(t *testing.T) {
	short := "abcdef"
	assert.Equal(t, short, TruncateDocText(short))

	long := strings.Repeat("x", MaxDocTextLen+1000)
	got := TruncateDocText(long)
	assert.Len(t, got, MaxDocTextLen)
	assert.Contains(t, got, "(document truncated)")
	assert.True(t, strings.HasPrefix(got, "xxxx"))
	assert.True(t, strings.HasSuffix(got, "xxxx"))
}
