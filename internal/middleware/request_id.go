package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hidgenerator"
)

const RequestIDKey = "request_id"

type requestIDContextKey struct{}

// NewRequestIDMiddleware 为请求链生成并传递唯一的 Request ID。
func NewRequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = hidgenerator.NewUuid().GenerateRequestID()
		}

		c.Request.Header.Set("X-Request-ID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Set(RequestIDKey, requestID)

		ctx := context.WithValue(c.Request.Context(), requestIDContextKey{}, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
