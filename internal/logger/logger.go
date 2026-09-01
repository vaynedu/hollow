package logger

import (
	"os"

	"github.com/vaynedu/hollow/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 创建日志实例。
func InitLogger(cfg *config.Config) (*zap.Logger, error) {
	var logConfig config.LogConfig
	if cfg != nil {
		logConfig = cfg.Log
	}

	// 设置默认值
	if logConfig.LogLevel == "" {
		logConfig.LogLevel = "debug"
	}
	if logConfig.OutputMode == "" {
		logConfig.OutputMode = "console"
	}
	if logConfig.LogFileName == "" {
		logConfig.LogFileName = "app.log"
	}
	if logConfig.MaxSize == 0 {
		logConfig.MaxSize = 100 // MB
	}
	if logConfig.MaxAge == 0 {
		logConfig.MaxAge = 30 // 天
	}

	var level zapcore.Level
	switch logConfig.LogLevel {
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
	if logConfig.OutputMode == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	var core zapcore.Core
	if logConfig.OutputMode == "file" {
		writer := &lumberjack.Logger{
			Filename:   logConfig.LogFileName,
			MaxSize:    logConfig.MaxSize,
			MaxBackups: 3,
			MaxAge:     logConfig.MaxAge,
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
