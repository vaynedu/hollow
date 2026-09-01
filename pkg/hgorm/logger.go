package hgorm

import (
	"context"
	"errors"
	"time"

	"github.com/vaynedu/hollow/pkg/hlog"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type contextLogger struct {
	slowThreshold time.Duration
	level         gormlogger.LogLevel
}

func newGORMLogger(slowThreshold time.Duration, level gormlogger.LogLevel) gormlogger.Interface {
	return &contextLogger{slowThreshold: slowThreshold, level: level}
}

func (l *contextLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *contextLogger) Info(ctx context.Context, message string, data ...any) {
	if l.level >= gormlogger.Info {
		hlog.FromContext(ctx).Sugar().Infof(message, data...)
	}
}

func (l *contextLogger) Warn(ctx context.Context, message string, data ...any) {
	if l.level >= gormlogger.Warn {
		hlog.FromContext(ctx).Sugar().Warnf(message, data...)
	}
}

func (l *contextLogger) Error(ctx context.Context, message string, data ...any) {
	if l.level >= gormlogger.Error {
		hlog.FromContext(ctx).Sugar().Errorf(message, data...)
	}
}

func (l *contextLogger) Trace(
	ctx context.Context,
	begin time.Time,
	query func() (sql string, rowsAffected int64),
	err error,
) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := query()
	fields := []hlog.Field{
		hlog.Duration("elapsed", elapsed),
		hlog.Int64("rows", rows),
		hlog.String("sql", sql),
	}
	logger := hlog.FromContext(ctx)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= gormlogger.Error:
		logger.Error("gorm sql error", append(fields, hlog.Err(err))...)
	case elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		logger.Warn("gorm slow sql", append(fields, hlog.Duration("threshold", l.slowThreshold))...)
	case l.level >= gormlogger.Info:
		logger.Info("gorm sql", fields...)
	}
}

func parseLogLevel(level string) gormlogger.LogLevel {
	switch level {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}
