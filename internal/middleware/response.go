package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaynedu/hollow/pkg/hecode"
	"go.uber.org/zap"
)

// NewResponseMiddleware 将处理结果封装为统一响应。
func NewResponseMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		requestID, _ := c.Get(RequestIDKey)
		requestIDValue := requestIDString(requestID)
		if ginErr := c.Errors.Last(); ginErr != nil {
			err := ginErr.Err
			var ecodeErr *hecode.EcodeError
			if !errors.As(err, &ecodeErr) {
				logger.Error("request failed",
					zap.Error(err),
					zap.String("request_id", requestIDValue),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				err = hecode.ErrInternal
			}

			status, response := hecode.Failure(requestIDValue, err)
			c.JSON(status, response)
			return
		}

		data, _ := c.Get("data")
		c.JSON(http.StatusOK, hecode.Success(requestIDValue, data))
	}
}

func requestIDString(value any) string {
	requestID, _ := value.(string)
	return requestID
}
