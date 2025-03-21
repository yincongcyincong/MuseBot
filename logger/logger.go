package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger instance
var Logger zerolog.Logger

// InitLogger init logger
func InitLogger() {
	fileWriter := &lumberjack.Logger{
		Filename:   "./log/telegram_deepseek.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   false,
	}

	stdoutWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	Logger = zerolog.New(zerolog.MultiLevelWriter(fileWriter, stdoutWriter)).With().
		Timestamp().
		Logger()

	log.SetOutput(Logger)
	log.SetFlags(0)
	// set log level
	switch strings.ToLower(*conf.LogLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func getCallerFile() string {
	_, filename, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", filename, line)
}

// Debug debug info
func Debug(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.Debug().Fields(fields).Msg(callerFile + " " + msg)
}

// Info info log
func Info(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.Info().Fields(fields).Msg(callerFile + " " + msg)
}

// Warn warn log
func Warn(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.Warn().Fields(fields).Msg(callerFile + " " + msg)
}

// Error error log
func Error(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.Error().Fields(fields).Msg(callerFile + " " + msg)
}

// Fatal fatal log
func Fatal(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.Fatal().Fields(fields).Msg(callerFile + " " + msg)
}
