package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware 异常恢复中间件
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err interface{}) {
		logger.Error("Panic recovered",
			zap.Any("error", err),
			zap.Stack("stack_trace"),
		)
		c.JSON(500, struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}{
			Code: 500,
			Msg:  "internal server error",
		})
	})
}
