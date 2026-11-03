package logger

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/grpclog"
)

// grpcLoggerV2 adapts grpc's logging interface onto MuseBot's logger.
// grpc-go writes transport level logs (used by the etcd, milvus and qdrant
// clients) through grpclog, whose default logger prints straight to os.Stderr
// and therefore never reaches muse_bot.log.
type grpcLoggerV2 struct{}

// grpc Info is chatty connectivity chatter, keep it under the debug level so
// the default info level stays readable; warnings and errors pass through.
func (g grpcLoggerV2) Info(args ...interface{}) {
	Logger.logger.Debug().Msg("grpc: " + fmt.Sprint(args...))
}

func (g grpcLoggerV2) Infoln(args ...interface{}) {
	Logger.logger.Debug().Msg("grpc: " + strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func (g grpcLoggerV2) Infof(format string, args ...interface{}) {
	Logger.logger.Debug().Msg("grpc: " + fmt.Sprintf(format, args...))
}

func (g grpcLoggerV2) Warning(args ...interface{}) {
	Logger.logger.Warn().Msg("grpc: " + fmt.Sprint(args...))
}

func (g grpcLoggerV2) Warningln(args ...interface{}) {
	Logger.logger.Warn().Msg("grpc: " + strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func (g grpcLoggerV2) Warningf(format string, args ...interface{}) {
	Logger.logger.Warn().Msg("grpc: " + fmt.Sprintf(format, args...))
}

func (g grpcLoggerV2) Error(args ...interface{}) {
	Logger.logger.Error().Msg("grpc: " + fmt.Sprint(args...))
}

func (g grpcLoggerV2) Errorln(args ...interface{}) {
	Logger.logger.Error().Msg("grpc: " + strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func (g grpcLoggerV2) Errorf(format string, args ...interface{}) {
	Logger.logger.Error().Msg("grpc: " + fmt.Sprintf(format, args...))
}

// zerolog's Fatal writes the event and calls os.Exit(1), which is exactly the
// semantics grpc's Fatalf expects from the logger.
func (g grpcLoggerV2) Fatal(args ...interface{}) {
	Logger.logger.Fatal().Msg("grpc: " + fmt.Sprint(args...))
}

func (g grpcLoggerV2) Fatalln(args ...interface{}) {
	Logger.logger.Fatal().Msg("grpc: " + strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
}

func (g grpcLoggerV2) Fatalf(format string, args ...interface{}) {
	Logger.logger.Fatal().Msg("grpc: " + fmt.Sprintf(format, args...))
}

func (g grpcLoggerV2) V(l int) bool {
	return l <= 0
}

// captureGrpcLog routes grpclog output into MuseBot's logger.
func captureGrpcLog() {
	grpclog.SetLoggerV2(grpcLoggerV2{})
}
