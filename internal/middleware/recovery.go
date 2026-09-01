package middleware

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hecode"
	"go.uber.org/zap"
)

// NewRecoveryMiddleware 捕获 panic 并交给统一响应中间件处理。
func NewRecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get(RequestIDKey)
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(stack())),
					zap.String("request_id", requestIDString(requestID)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				_ = c.Error(hecode.ErrInternal)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// stack 返回当前 goroutine 的堆栈信息。
func stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
