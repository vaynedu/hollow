package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hidgenerator"
)

const RequestIDKey = "request_id"

// RequestIDMiddleware 实现Middleware接口的请求ID中间件
type RequestIDMiddleware struct{}

// NewRequestIDMiddleware 创建RequestIDMiddleware实例
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// HandlerFunc 返回中间件处理函数
func (m *RequestIDMiddleware) HandlerFunc() gin.HandlerFunc {
	return m.requestIDMiddleware
}

// Identifier 返回中间件唯一标识
func (m *RequestIDMiddleware) Identifier() string {
	return "request_id"
}

func (m *RequestIDMiddleware) requestIDMiddleware(c *gin.Context) {
	// 从请求头中获取 request_id
	requestID := c.GetHeader("X-Request-ID")

	// 如果请求头中没有 request_id，则生成一个新的
	if requestID == "" {
		requestID = hidgenerator.NewUuid().GenerateRequestID()
		// 将新生成的 request_id 添加到请求头中
		c.Request.Header.Set("X-Request-ID", requestID)
	}

	// 将 request_id 添加到响应头中
	c.Writer.Header().Set("X-Request-ID", requestID)

	// 将 request_id 保存到 context 中
	ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
	c.Request = c.Request.WithContext(ctx)

	c.Next()
}
