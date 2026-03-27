package middleware

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/internal/logger"
	"go.uber.org/zap"
)

// RecoveryMiddleware 实现Middleware接口的恢复中间件
type RecoveryMiddleware struct {
	logger *zap.Logger
}

// NewRecoveryMiddleware 创建RecoveryMiddleware实例
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger.GetLogger(),
	}
}

// HandlerFunc 返回中间件处理函数
func (m *RecoveryMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印错误堆栈
				stack := stack(3)
				m.logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(stack)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// 返回500响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// Identifier 返回中间件唯一标识
func (m *RecoveryMiddleware) Identifier() string {
	return "recovery"
}

// stack returns a formatted stack trace of the goroutine that calls it.
func stack(skip int) []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
