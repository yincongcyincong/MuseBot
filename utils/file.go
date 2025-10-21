package utils

import (
	"bytes"
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

func GetTailStartOffset(filePath string, lines int) (int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	const bufferSize = 4096
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	
	size := stat.Size()
	var offset = size
	var count int
	
	for offset > 0 && count <= lines {
		readSize := int64(bufferSize)
		if offset < readSize {
			readSize = offset
		}
		offset -= readSize
		tmp := make([]byte, readSize)
		if _, err := file.ReadAt(tmp, offset); err != nil {
			return 0, err
		}
		count += bytes.Count(tmp, []byte("\n"))
	}
	
	if offset <= 0 {
		offset = 0
	}
	
	return offset, nil
}
