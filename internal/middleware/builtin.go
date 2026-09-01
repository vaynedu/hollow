package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterDefaultMiddlewares 注册默认的中间件
func RegisterDefaultMiddlewares(logger *zap.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		NewRequestIDMiddleware(),
		NewLoggingMiddleware(logger),
		NewResponseMiddleware(logger),
		NewRecoveryMiddleware(logger),
	}
}
