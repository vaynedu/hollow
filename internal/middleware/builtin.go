package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hlog"
	"go.uber.org/zap"
)

// RegisterDefaultMiddlewares 注册默认的中间件
func RegisterDefaultMiddlewares(logger *zap.Logger) []gin.HandlerFunc {
	hlog.SetDefault(logger)
	return []gin.HandlerFunc{
		NewRequestIDMiddleware(),
		NewLoggingMiddleware(logger),
		NewResponseMiddleware(logger),
		NewRecoveryMiddleware(logger),
	}
}
