package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MetricsMiddleware 实现Middleware接口的 metrics 中间件
type MetricsMiddleware struct {
	logger *zap.Logger
}

// NewMetricsMiddleware 创建MetricsMiddleware实例
func NewMetricsMiddleware(logger *zap.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{logger: logger}
}

// HandlerFunc 返回中间件处理函数
func (m *MetricsMiddleware) HandlerFunc() gin.HandlerFunc {
	return m.metricsMiddleware
}

// Identifier 返回中间件唯一标识
func (m *MetricsMiddleware) Identifier() string {
	return "metrics"
}

func (m *MetricsMiddleware) metricsMiddleware(c *gin.Context) {
	start := time.Now()

	c.Next()

	duration := time.Since(start)
	// 这里可以添加指标收集逻辑，例如请求耗时、状态码等
	m.logger.Info("请求耗时",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Duration("duration", duration),
	)
}
