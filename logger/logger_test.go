package logger

import (
	"context"
	"testing"
)

func TestLogger(t *testing.T) {
	InitLogger()

	Info("Test Info log", map[string]interface{}{"test": "info"})
	Warn("Test Warn log", map[string]interface{}{"test": "warn"})
	Error("Test Error log", map[string]interface{}{"test": "error"})

	// Debug 需要手动设置 log level，否则可能不输出
	Debug("Test Debug log", map[string]interface{}{"test": "debug"})
}

func TestLoggerMethods(t *testing.T) {

	ctx := context.Background()

	Logger.Debug(ctx, map[string]interface{}{"test": "debug"})
	Logger.Info(ctx, map[string]interface{}{"test": "info"})
	Logger.Warn(ctx, map[string]interface{}{"test": "warn"})
	Logger.Error(ctx, map[string]interface{}{"test": "error"})

	Logger.Debugf("Debugf test: %s", "debug")
	Logger.Infof("Infof test: %s", "info")
	Logger.Warningf("Warningf test: %s", "warn")
	Logger.Errorf("Errorf test: %s", "error")

}

func TestColorFormatLevel(t *testing.T) {
	cases := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL", "PANIC", "OTHER"}
	for _, level := range cases {
		colored := Logger.ColorFormatLevel(level)
		if colored == "" {
			t.Errorf("ColorFormatLevel returned empty for %s", level)
		}
	}
}
