package hlog

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// ErrInvalidLevel 表示日志级别配置不受支持。
	ErrInvalidLevel = errors.New("hlog: invalid level")
	// ErrInvalidOutputMode 表示日志输出模式配置不受支持。
	ErrInvalidOutputMode = errors.New("hlog: invalid output mode")
)

// Config 定义结构化日志和文件轮转配置。
type Config struct {
	Level      string `mapstructure:"level"`
	OutputMode string `mapstructure:"output_mode"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// New 根据配置创建 Logger。
func New(cfg Config) (*Logger, error) {
	applyDefaults(&cfg)
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLevel, cfg.Level)
	}
	if cfg.OutputMode != "console" && cfg.OutputMode != "file" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidOutputMode, cfg.OutputMode)
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

	var core zapcore.Core
	if cfg.OutputMode == "file" {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
		core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, level)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), level)
	}

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)), nil
}

func applyDefaults(cfg *Config) {
	if cfg.Level == "" {
		cfg.Level = "debug"
	}
	if cfg.OutputMode == "" {
		cfg.OutputMode = "console"
	}
	if cfg.File == "" {
		cfg.File = "app.log"
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 3
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 30
	}
}
