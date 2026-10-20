package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yincongcyincong/MuseBot/logger"
)

const MediaRootDir = "data/media"

// SanitizeChatId keeps only path-safe characters so a chat id (or a
// user-controlled web user id) can never escape its media folder.
func SanitizeChatId(chatId string) string {
	var b strings.Builder
	for _, r := range chatId {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		if b.Len() >= 64 {
			break
		}
	}
	return b.String()
}

func MediaAbsDir() string {
	return GetAbsPath(MediaRootDir)
}

func IsMediaPath(p string) bool {
	return strings.HasPrefix(p, MediaRootDir+"/")
}

// SaveMediaFile writes content into data/media/{folder}/ and returns the
// relative path stored in db records. folder must already be sanitized
// (see SanitizeChatId), origName may be empty.
func SaveMediaFile(folder, origName string, content []byte, ext string) (string, error) {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	base := sanitizeFileBase(origName)
	if base != "" {
		base += "_"
	}
	fileName := base + RandomFilename(ext)

	absDir := filepath.Join(MediaAbsDir(), folder)
	err := os.MkdirAll(absDir, 0755)
	if err != nil {
		return "", fmt.Errorf("mkdir media dir fail: %w", err)
	}

	absPath := filepath.Join(absDir, fileName)
	err = os.WriteFile(absPath, content, 0644)
	if err != nil {
		return "", fmt.Errorf("write media file fail: %w", err)
	}

	return MediaRootDir + "/" + folder + "/" + fileName, nil
}

func sanitizeFileBase(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		}
		if b.Len() >= 64 {
			break
		}
	}
	return b.String()
}

// ReadMediaFile reads back a media file by its relative path stored in db.
func ReadMediaFile(relPath string) ([]byte, error) {
	if !IsMediaPath(relPath) || strings.Contains(relPath, "..") {
		return nil, fmt.Errorf("invalid media path: %s", relPath)
	}
	return os.ReadFile(GetAbsPath(relPath))
}

// StartMediaRetention launches the daily media cleanup goroutine.
// days <= 0 disables cleanup.
func StartMediaRetention(days int) {
	if days <= 0 {
		return
	}

	go func() {
		RunMediaRetention(days)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			RunMediaRetention(days)
		}
	}()
}

// RunMediaRetention deletes media files older than days, then removes the
// empty conversation folders left behind.
func RunMediaRetention(days int) {
	if days <= 0 {
		return
	}
	if _, err := os.Stat(MediaAbsDir()); err != nil {
		// nothing stored yet
		return
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	var folders []string

	err := filepath.WalkDir(MediaAbsDir(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != MediaAbsDir() {
				// WalkDir yields parents before children, so reversing
				// later removes leaf folders first.
				folders = append(folders, path)
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			logger.Warn("get media file info fail", "file", path, "err", err)
			return nil
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				logger.Warn("remove expired media file fail", "file", path, "err", err)
			}
		}
		return nil
	})
	if err != nil {
		logger.Warn("walk media dir fail", "err", err)
		return
	}

	// reversed: deepest first so emptied leaf folders can empty their parents
	for i := len(folders) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(folders[i])
		if err == nil && len(entries) == 0 {
			_ = os.Remove(folders[i])
		}
	}
}
