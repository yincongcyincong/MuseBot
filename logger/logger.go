package logger

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	
	"github.com/mgutz/ansi"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerInfo struct {
	logger zerolog.Logger
}

type QQLoggerInfo struct {
	logger zerolog.Logger
}

var (
	LogLevel *string
)

func init() {
	LogLevel = flag.String("log_level", "info", "log level")
	
	if os.Getenv("LOG_LEVEL") != "" {
		*LogLevel = os.Getenv("LOG_LEVEL")
	}
	
	fmt.Println("log level:", *LogLevel)
}

// Logger instance
var Logger = new(LoggerInfo)
var QQLogger = new(QQLoggerInfo)

// InitLogger init logger
func InitLogger() {
	fileWriter := &lumberjack.Logger{
		Filename:   "./log/muse_bot.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   false,
	}
	
	stdoutWriter := zerolog.ConsoleWriter{
		Out:         os.Stdout,
		TimeFormat:  "2006-01-02 15:04:05",
		FormatLevel: Logger.ColorFormatLevel,
	}
	
	Logger.logger = zerolog.New(zerolog.MultiLevelWriter(fileWriter, stdoutWriter)).With().
		Timestamp().
		Logger()
	QQLogger.logger = Logger.logger
	
	log.SetOutput(Logger.logger)
	log.SetFlags(0)
	// set log level
	switch strings.ToLower(*LogLevel) {
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
	
	Info("log level", "loglevel", *LogLevel)
}

func getCallerFile() string {
	_, filename, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", filename, line)
}

// Debug debug info
func (l *LoggerInfo) Debug(ctx context.Context, fields ...interface{}) {
	callerFile := getCallerFile()
	zerolog.Ctx(ctx).Debug().Fields(fields).Msg(callerFile)
}

// Info info log
func (l *LoggerInfo) Info(ctx context.Context, fields ...interface{}) {
	callerFile := getCallerFile()
	zerolog.Ctx(ctx).Info().Fields(fields).Msg(callerFile)
}

func (l *LoggerInfo) Output(level int, logStr string) error {
	callerFile := getCallerFile()
	l.logger.Info().Msg(callerFile + " " + logStr)
	return nil
}

// Warn log
func (l *LoggerInfo) Warn(ctx context.Context, fields ...interface{}) {
	callerFile := getCallerFile()
	zerolog.Ctx(ctx).Warn().Fields(fields).Msg(callerFile)
}

// Error error log
func (l *LoggerInfo) Error(ctx context.Context, fields ...interface{}) {
	callerFile := getCallerFile()
	zerolog.Ctx(ctx).Error().Fields(fields).Msg(callerFile)
}

// Fatal fatal log
func (l *LoggerInfo) Fatal(ctx context.Context, fields ...interface{}) {
	callerFile := getCallerFile()
	zerolog.Ctx(ctx).Fatal().Fields(fields).Msg(callerFile)
}

// Debugf debug log
func (l *LoggerInfo) Debugf(format string, args ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Debug().Msg(callerFile + " " + fmt.Sprintf(format, args...))
}

// Infof info log
func (l *LoggerInfo) Infof(format string, args ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Info().Msg(callerFile + " " + fmt.Sprintf(format, args...))
}

// Warningf log
func (l *LoggerInfo) Warningf(format string, args ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Warn().Msg(callerFile + " " + fmt.Sprintf(format, args...))
}

// Errorf error log
func (l *LoggerInfo) Errorf(format string, args ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Error().Msg(callerFile + " " + fmt.Sprintf(format, args...))
}

// Fatalf fatal log
func (l *LoggerInfo) Fatalf(format string, args ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Fatal().Msg(callerFile + " " + fmt.Sprintf(format, args...))
}

func (l *LoggerInfo) ColorFormatLevel(i interface{}) string {
	level := strings.ToUpper(fmt.Sprintf("%v", i))
	switch level {
	case "DEBUG":
		return ansi.Color(fmt.Sprintf("| %-5s |", level), "cyan")
	case "INFO":
		return ansi.Color(fmt.Sprintf("| %-5s |", level), "green")
	case "WARN":
		return ansi.Color(fmt.Sprintf("| %-5s |", level), "yellow")
	case "ERROR":
		return ansi.Color(fmt.Sprintf("| %-5s |", level), "red")
	case "FATAL":
		return ansi.Color(fmt.Sprintf("| %-5s |", level), "magenta")
	case "PANIC":
		return ansi.Color(fmt.Sprintf("| %-5s |", level), "magenta+bh")
	default:
		return ansi.Color(fmt.Sprintf("| %-5s |", "STD"), "white")
	}
}

func Debug(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Debug().Fields(fields).Msg(callerFile + " " + msg)
}

// Info info log
func Info(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Info().Fields(fields).Msg(callerFile + " " + msg)
}

// Warn log
func Warn(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Warn().Fields(fields).Msg(callerFile + " " + msg)
}

// Error error log
func Error(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Error().Fields(fields).Msg(callerFile + " " + msg)
}

// Fatal fatal log
func Fatal(msg string, fields ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Fatal().Fields(fields).Msg(callerFile + " " + msg)
}

func (l *QQLoggerInfo) Debug(v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Debug().Fields(v).Msg(callerFile)
}
func (l *QQLoggerInfo) Info(v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Info().Fields(v).Msg(callerFile)
}
func (l *QQLoggerInfo) Warn(v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Warn().Fields(v).Msg(callerFile)
}
func (l *QQLoggerInfo) Error(v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Error().Fields(v).Msg(callerFile)
}

func (l *QQLoggerInfo) Debugf(format string, v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Debug().Msg(callerFile + " " + fmt.Sprintf(format, v...))
}
func (l *QQLoggerInfo) Infof(format string, v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Info().Msg(callerFile + " " + fmt.Sprintf(format, v...))
}
func (l *QQLoggerInfo) Warnf(format string, v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Warn().Msg(callerFile + " " + fmt.Sprintf(format, v...))
}
func (l *QQLoggerInfo) Errorf(format string, v ...interface{}) {
	callerFile := getCallerFile()
	Logger.logger.Error().Msg(callerFile + " " + fmt.Sprintf(format, v...))
}

// Sync logger Sync calls to flush buffer
func (l *QQLoggerInfo) Sync() error {
	return nil
}
