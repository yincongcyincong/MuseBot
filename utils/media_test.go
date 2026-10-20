package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeChatId(t *testing.T) {
	assert.Equal(t, "telegram_123", SanitizeChatId("telegram_123"))
	assert.Equal(t, "C01_ABC_def", SanitizeChatId("C01/ABC:def"))
	assert.Equal(t, "__", SanitizeChatId(".."))
	assert.Equal(t, "", SanitizeChatId(""))

	long := strings.Repeat("a", 100)
	assert.Len(t, SanitizeChatId(long), 64)

	// path separators can never survive into a folder name
	assert.NotContains(t, SanitizeChatId("a/b\\c"), "/")
	assert.NotContains(t, SanitizeChatId("a/b\\c"), "\\")
}

func TestSaveAndReadMediaFile(t *testing.T) {
	content := []byte("hello media")

	relPath, err := SaveMediaFile("web_u1", "报告.pdf", content, "pdf")
	assert.NoError(t, err)
	assert.True(t, IsMediaPath(relPath))
	assert.Contains(t, relPath, "data/media/web_u1/")
	assert.Contains(t, filepath.Base(relPath), ".pdf")

	data, err := ReadMediaFile(relPath)
	assert.NoError(t, err)
	assert.Equal(t, content, data)

	// traversal and outside-root reads are rejected
	_, err = ReadMediaFile("data/media/../muse_bot.db")
	assert.Error(t, err)
	_, err = ReadMediaFile("etc/passwd")
	assert.Error(t, err)
}

func TestRunMediaRetention(t *testing.T) {
	old, err := SaveMediaFile("web_old", "old.txt", []byte("old"), "txt")
	assert.NoError(t, err)
	fresh, err := SaveMediaFile("web_fresh", "fresh.txt", []byte("fresh"), "txt")
	assert.NoError(t, err)

	past := time.Now().AddDate(0, 0, -40)
	assert.NoError(t, os.Chtimes(GetAbsPath(old), past, past))

	RunMediaRetention(30)

	_, err = os.Stat(GetAbsPath(old))
	assert.True(t, os.IsNotExist(err), "expired file should be removed")

	_, err = os.Stat(GetAbsPath(fresh))
	assert.NoError(t, err, "fresh file should survive")

	// days=0 never cleans
	RunMediaRetention(0)
	_, err = os.Stat(GetAbsPath(fresh))
	assert.NoError(t, err)
}

func TestDetectDocumentType(t *testing.T) {
	assert.Equal(t, ".pdf", DetectDocumentType("report.PDF", nil))
	assert.Equal(t, ".txt", DetectDocumentType("notes.txt", nil))
	assert.Equal(t, ".htm", DetectDocumentType("page.htm", nil))
	assert.Equal(t, ".csv", DetectDocumentType("data.csv", nil))

	// sniffing without a usable file name
	assert.Equal(t, ".pdf", DetectDocumentType("", []byte("%PDF-1.7 fake")))
	assert.Equal(t, ".html", DetectDocumentType("", []byte("<!DOCTYPE html><p>x")))
	assert.Equal(t, ".html", DetectDocumentType("", []byte("<html><body>")))

	// unknown extensions fall back to plain text when the content is text
	assert.Equal(t, ".txt", DetectDocumentType("deploy.bat", []byte("@echo off\r\necho hi\r\n")))
	assert.Equal(t, ".txt", DetectDocumentType("changes.diff", []byte("--- a/x\r\n+++ b/x\r\n")))
	assert.Equal(t, ".txt", DetectDocumentType("", []byte("plain bytes")))

	// an empty body never claims to be a document (guards plain web requests)
	assert.Equal(t, "", DetectDocumentType("", nil))
	assert.Equal(t, "", DetectDocumentType("song.mp3", nil))

	// binary content is rejected
	assert.Equal(t, "", DetectDocumentType("archive.zip", []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x01, 0x02, 0x00}))
	assert.Equal(t, "", DetectDocumentType("", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0xFF, 0xD9}))
}

func TestLooksLikeText(t *testing.T) {
	assert.True(t, LooksLikeText(nil))
	assert.True(t, LooksLikeText([]byte("hello")))
	assert.True(t, LooksLikeText([]byte("中文文本")))
	// mostly-printable non-UTF-8 (GBK) still counts as text
	assert.True(t, LooksLikeText([]byte{0xD6, 0xD0, 0xCE, 0xC4, '\n', 'a', 'b'}))
	// control-char dominated bytes are binary
	assert.False(t, LooksLikeText([]byte{0x00, 0x01, 0x02, 0x03, 'a', 'b'}))
}
