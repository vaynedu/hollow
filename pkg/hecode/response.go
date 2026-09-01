package hecode

import (
	"errors"
	"net/http"
)

// Response 标准响应结构
type Response struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	RequestID string `json:"request_id"`
	Data      any    `json:"data"`
}

// Success 返回成功响应
func Success(requestID string, data any) Response {
	return Response{Code: 0, Msg: "success", RequestID: requestID, Data: data}
}

// Failure 返回错误响应及 HTTP 状态码
func Failure(requestID string, err error) (int, Response) {
	status, code, message := Resolve(err)
	return status, Response{Code: code, Msg: message, RequestID: requestID, Data: nil}
}

// Resolve 将业务错误转换为 HTTP 状态码、业务码和安全消息
func Resolve(err error) (int, int, string) {
	var target *EcodeError
	if !errors.As(err, &target) {
		return http.StatusInternalServerError, Code(ErrInternal), "internal server error"
	}
	status := http.StatusInternalServerError
	switch target.Code() {
	case 1100:
		status = http.StatusBadRequest
	case 1204:
		status = http.StatusUnauthorized
	case 1202, 1203:
		status = http.StatusForbidden
	case 1200:
		status = http.StatusNotFound
	}
	return status, target.Code(), target.GetMessage()
}
