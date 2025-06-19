package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterDefaultMiddlewares 注册默认的中间件
func RegisterDefaultMiddlewares(logger *zap.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		LoggingMiddleware(logger),  // 日志中间件
		RecoveryMiddleware(logger), // 恢复中间件
		ResponseMiddleware(),       // 响应中间件
		//RequestIDMiddleware(),      // 请求ID中间件
	}
}
