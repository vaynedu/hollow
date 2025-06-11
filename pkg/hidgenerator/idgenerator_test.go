package hidgenerator

import (
	"fmt"
	"testing"
)

func TestGenerateID(t *testing.T) {

	var generator IdGenerator = NewUuid()

	// 调用 GenerateRequestID 方法
	requestID := generator.GenerateRequestID()
	fmt.Printf("生成的请求 ID: %s\n", requestID)
}
