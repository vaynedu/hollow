package hlog

import "go.uber.org/zap"

var (
	Any      = zap.Any
	Bool     = zap.Bool
	Duration = zap.Duration
	Int      = zap.Int
	Int32    = zap.Int32
	Int64    = zap.Int64
	String   = zap.String
	Strings  = zap.Strings
	Time     = zap.Time
	Uint64   = zap.Uint64
)

// Err 创建标准 error 日志字段。
func Err(err error) Field {
	return zap.Error(err)
}
