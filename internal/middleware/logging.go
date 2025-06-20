package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

// LoggingMiddleware 实现Middleware接口的日志中间件
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware 创建LoggingMiddleware实例
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// HandlerFunc 返回中间件处理函数
func (m *LoggingMiddleware) HandlerFunc() gin.HandlerFunc {
	return m.loggingMiddleware
}

// Identifier 返回中间件唯一标识
func (m *LoggingMiddleware) Identifier() string {
	return "logging"
}

func (m *LoggingMiddleware) loggingMiddleware(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	c.Next()

	cost := time.Since(start)
	m.logger.Info("HTTP Request",
		zap.String("method", c.Request.Method),
		zap.String("path", path),
		zap.String("query", query),
		zap.Int("status", c.Writer.Status()),
		zap.Duration("cost", cost),
		zap.String("client_ip", c.ClientIP()),
	)
}
