package logger

import (
	"fmt"
	"os"

	"github.com/vaynedu/hollow/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

// GetLogger 获取全局日志实例
func GetLogger() *zap.Logger {
	if logger == nil {
		InitLogger(nil)
	}
	return logger
}

// InitLogger 初始化日志
func InitLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg == nil {
		cfg = &config.Config{
			Log: config.LogConfig{},
		}
	}

	// 设置默认值
	if cfg.Log.LogLevel == "" {
		cfg.Log.LogLevel = "debug"
	}
	if cfg.Log.OutputMode == "" {
		cfg.Log.OutputMode = "console"
	}
	if cfg.Log.LogFileName == "" {
		cfg.Log.LogFileName = "app.log"
	}
	if cfg.Log.MaxSize == 0 {
		cfg.Log.MaxSize = 100 // MB
	}
	if cfg.Log.MaxAge == 0 {
		cfg.Log.MaxAge = 30 // 天
	}

	var level zapcore.Level
	switch cfg.Log.LogLevel {
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
	if cfg.Log.OutputMode == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	var core zapcore.Core
	if cfg.Log.OutputMode == "file" {
		writer := &lumberjack.Logger{
			Filename:   cfg.Log.LogFileName,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: 3,
			MaxAge:     cfg.Log.MaxAge,
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

	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	return logger, nil
}

// Debug 打印 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 打印 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 打印 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 打印 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 打印 Fatal 级别日志并退出
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Debugf 格式化打印 Debug 级别日志
func Debugf(format string, args ...interface{}) {
	GetLogger().Debug(fmt.Sprintf(format, args...))
}

// Infof 格式化打印 Info 级别日志
func Infof(format string, args ...interface{}) {
	GetLogger().Info(fmt.Sprintf(format, args...))
}

// Warnf 格式化打印 Warn 级别日志
func Warnf(format string, args ...interface{}) {
	GetLogger().Warn(fmt.Sprintf(format, args...))
}

// Errorf 格式化打印 Error 级别日志
func Errorf(format string, args ...interface{}) {
	GetLogger().Error(fmt.Sprintf(format, args...))
}

// Fatalf 格式化打印 Fatal 级别日志并退出
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatal(fmt.Sprintf(format, args...))
}

// WithFields 创建带字段的日志实例
func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	return GetLogger().Sync()
}
