package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {
	InitLogger()

	Info("Test Info log", map[string]interface{}{"test": "info"})
	Warn("Test Warn log", map[string]interface{}{"test": "warn"})
	Error("Test Error log", map[string]interface{}{"test": "error"})
}
