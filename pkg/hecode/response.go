package hecode

// Response 标准响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success 返回成功响应
func Success(data interface{}) Response {
	return Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
}

// Error 返回错误响应
func Error(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
	}
}

// ErrorWithData 返回带数据的错误响应
func ErrorWithData(code int, msg string, data interface{}) Response {
	return Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}
