package utils

import (
	"os"
	"path/filepath"

	"github.com/yincongcyincong/MuseBot/logger"
)

func GetAbsPath(relPath string) string {
	exe, err := os.Executable()
	if err != nil {
		logger.Error("Failed to get executable path", "err", err)
		return ""
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, relPath)
}
