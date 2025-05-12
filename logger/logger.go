package logger

import (
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
        Out:         os.Stdout,
        TimeFormat:  "2006-01-02 15:04:05",
        FormatLevel: ColorFormatLevel,
    }

    Logger = zerolog.New(zerolog.MultiLevelWriter(fileWriter, stdoutWriter)).With().
        Timestamp().
        Logger()

    log.SetOutput(Logger)
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
func Debug(msg string, fields ...interface{}) {
    callerFile := getCallerFile()
    Logger.Debug().Fields(fields).Msg(callerFile + " " + msg)
}

// Info info log
func Info(msg string, fields ...interface{}) {
    callerFile := getCallerFile()
    Logger.Info().Fields(fields).Msg(callerFile + " " + msg)
}

// Warn log
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

func ColorFormatLevel(i interface{}) string {
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
