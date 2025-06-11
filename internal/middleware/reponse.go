package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hidgenerator"
)

// Response 标准响应格式
type Response struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
}

// ResponseMiddleware 中间件：自动封装响应
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 从 context 中获取 request_id
		requestId, exists := c.Get(RequestIDKey)
		if !exists {
			requestId = hidgenerator.NewUuid().GenerateRequestID()
			// 将新生成的 request_id 添加到请求头中
			c.Request.Header.Set("X-Request-ID", requestId.(string))
		}

		// 处理业务错误
		if err := c.Errors.Last(); err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Code:      http.StatusInternalServerError,
				Msg:       err.Err.Error(),
				RequestID: requestId.(string),
			})
			return
		}

		// 获取业务数据（由handler通过c.Set("data", ...)设置）
		data, exists := c.Get("data")
		if !exists {
			data = nil
		}

		c.JSON(http.StatusOK, Response{
			Code:      http.StatusOK,
			Msg:       "success",
			RequestID: requestId.(string),
			Data:      data,
		})
	}
}
