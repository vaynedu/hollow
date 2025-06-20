package middleware

import (
	"go.uber.org/zap"
)

// RegisterDefaultMiddlewares 注册默认的中间件
func RegisterDefaultMiddlewares(logger *zap.Logger) []Middleware {
	return []Middleware{
		NewRequestIDMiddleware(),     // 请求ID中间件
		NewLoggingMiddleware(logger), // 日志中间件
		NewRecoveryMiddleware(),      // 恢复中间件
		// NewMetricsMiddleware(), // metrics 中间件
		NewResponseMiddleware(), // 响应中间件
	}
}
