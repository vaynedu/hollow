package hlog

import (
	"sync/atomic"

	"go.uber.org/zap"
)

var global atomic.Pointer[zap.Logger]

func init() {
	global.Store(zap.NewNop())
}

// SetDefault 替换进程默认 Logger；nil 会恢复为无输出 Logger。
func SetDefault(logger *Logger) {
	if logger == nil {
		logger = zap.NewNop()
	}
	global.Store(logger)
}

// L 返回进程默认 Logger。
func L() *Logger {
	return global.Load()
}

// Logger 是 zap.Logger 的别名。
type Logger = zap.Logger

// Field 是 zap.Field 的别名。
type Field = zap.Field
