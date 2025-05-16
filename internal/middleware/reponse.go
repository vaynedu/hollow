package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 标准响应格式
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// ResponseMiddleware 中间件：自动封装响应
func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理业务错误
		if err := c.Errors.Last(); err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Code: http.StatusInternalServerError,
				Msg:  err.Err.Error(),
			})
			return
		}

		// 获取业务数据（由handler通过c.Set("data", ...)设置）
		data, exists := c.Get("data")
		if !exists {
			data = nil
		}

		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  "success",
			Data: data,
		})
	}
}
