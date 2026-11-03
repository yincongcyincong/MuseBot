package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
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

// TestCtxMethodsEmitEvents guards the regression where the ctx variants went
// through zerolog.Ctx(ctx), which returns a disabled logger for SDK contexts
// and silently dropped everything.
func TestCtxMethodsEmitEvents(t *testing.T) {
	InitLogger()

	var buf bytes.Buffer
	orig := Logger.logger
	defer func() { Logger.logger = orig }()
	Logger.logger = zerolog.New(&buf)

	ctx := context.WithValue(context.Background(), "bot_name", "bot1")
	ctx = context.WithValue(ctx, "log_id", "lid1")

	// SDK call shape: plain message args, possibly an odd count
	Logger.Info(ctx, "ws connected", "[conn_id=abc]")
	out := buf.String()
	if !strings.Contains(out, `"bot_name":"bot1"`) || !strings.Contains(out, `"log_id":"lid1"`) {
		t.Fatalf("ctx fields missing: %s", out)
	}
	if !strings.Contains(out, "ws connected") || !strings.Contains(out, "conn_id=abc") {
		t.Fatalf("sdk args lost: %s", out)
	}

	// an odd arg count must not lose the trailing arg
	buf.Reset()
	Logger.Warn(ctx, "single arg")
	if !strings.Contains(buf.String(), "single arg") {
		t.Fatalf("odd arg count dropped: %s", buf.String())
	}

	// nil ctx (some SDKs pass it) must not panic
	Logger.Error(nil, "nil ctx message")
	if !strings.Contains(buf.String(), "nil ctx message") {
		t.Fatalf("nil ctx log dropped: %s", buf.String())
	}
}

func TestGrpcLoggerV2(t *testing.T) {
	InitLogger()

	var buf bytes.Buffer
	orig := Logger.logger
	defer func() { Logger.logger = orig }()
	Logger.logger = zerolog.New(&buf)

	g := grpcLoggerV2{}
	g.Warningf("conn %s", "down")
	g.Errorf("dial %s", "refused")
	g.Warningln("a", "b")

	out := buf.String()
	if !strings.Contains(out, `"level":"warn"`) || !strings.Contains(out, "grpc: conn down") {
		t.Fatalf("grpc warning missing: %s", out)
	}
	if !strings.Contains(out, `"level":"error"`) || !strings.Contains(out, "grpc: dial refused") {
		t.Fatalf("grpc error missing: %s", out)
	}
	// Infoln must trim its trailing newline, not smuggle it into the message
	if !strings.Contains(out, "grpc: a b") || strings.Contains(out, `a b\n`) {
		t.Fatalf("grpc Infoln formatting wrong: %s", out)
	}
	// grpc info is demoted to debug, invisible at the default level
	if strings.Contains(out, "grpc info") {
		t.Fatalf("grpc info should be debug-level: %s", out)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	buf.Reset()
	g.Infof("grpc info %d", 1)
	if !strings.Contains(buf.String(), `"level":"debug"`) || !strings.Contains(buf.String(), "grpc: grpc info 1") {
		t.Fatalf("grpc info event missing at debug level: %s", buf.String())
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
