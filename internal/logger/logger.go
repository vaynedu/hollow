package logger

import (
	"os"

	"github.com/vaynedu/hollow/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(cfg *config.Config) (*zap.Logger, error) {
	// 设置默认值
	if cfg.LogConfig.LogLevel == "" {
		cfg.LogConfig.LogLevel = "debug"
	}
	if cfg.LogConfig.OutputMode == "" {
		cfg.LogConfig.OutputMode = "console"
	}
	if cfg.LogConfig.LogFileName == "" {
		cfg.LogConfig.LogFileName = "app.log"
	}
	if cfg.LogConfig.MaxSize == 0 {
		cfg.LogConfig.MaxSize = 100 // MB
	}
	if cfg.LogConfig.MaxAge == 0 {
		cfg.LogConfig.MaxAge = 30 // 天
	}

	var level zapcore.Level
	switch cfg.LogConfig.LogLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.DebugLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if cfg.LogConfig.OutputMode == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	var core zapcore.Core
	if cfg.LogConfig.OutputMode == "file" {
		writer := &lumberjack.Logger{
			Filename:   cfg.LogConfig.LogFileName,
			MaxSize:    cfg.LogConfig.MaxSize,
			MaxBackups: 3,
			MaxAge:     cfg.LogConfig.MaxAge,
			Compress:   false,
		}
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(writer),
			level,
		)
	} else {
		consoleWriter := zapcore.AddSync(os.Stdout)
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(consoleWriter),
			level,
		)
	}

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)), nil
}
