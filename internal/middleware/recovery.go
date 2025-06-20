package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime"
)

// RecoveryMiddleware 实现Middleware接口的恢复中间件
type RecoveryMiddleware struct{}

// NewRecoveryMiddleware 创建RecoveryMiddleware实例
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{}
}

// HandlerFunc 返回中间件处理函数
func (m *RecoveryMiddleware) HandlerFunc() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, m.recoveryMiddleware)
}

// Identifier 返回中间件唯一标识
func (m *RecoveryMiddleware) Identifier() string {
	return "recovery"
}

func (m *RecoveryMiddleware) recoveryMiddleware(c *gin.Context, err interface{}) {
	if err != nil {
		// 打印错误堆栈
		stack := stack(3)
		log.Printf("panic recovered:\n%v\n%s", err, stack)

		// 返回500响应
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Internal Server Error",
		})
	}
}

// stack returns a formatted stack trace of the goroutine that calls it.
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
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
