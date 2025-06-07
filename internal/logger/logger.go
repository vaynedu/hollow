package logger

import (
	"github.com/vaynedu/hollow/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(cfg *config.Config) (*zap.Logger, error) {
	logCfg := zap.NewProductionConfig()
	logCfg.EncoderConfig.TimeKey = "timestamp"
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logCfg.EncoderConfig.StacktraceKey = "stacktrace"
	logCfg.EncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	logCfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	return logCfg.Build()
}
