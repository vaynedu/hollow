package hutils

import (
	"github.com/google/uuid"
)

// GenerateRequestID 生成请求ID，唯一标识符
func GenerateRequestID() string {
	return uuid.NewString()
}
