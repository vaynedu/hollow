package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hlog"
	"go.uber.org/zap"
)

// NewLoggingMiddleware 记录 HTTP 请求的基本信息。
func NewLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		requestLogger := hlog.FromContext(c.Request.Context())
		requestLogger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("cost", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}
