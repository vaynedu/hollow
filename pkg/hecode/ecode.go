package hecode

import (
	"fmt"
	"sync"
)

// 全局变量，用于跟踪已使用的错误码
var (
	usedCodes     = make(map[int]bool) // 存储已使用的错误码
	codesMutex    = &sync.RWMutex{}    // 用于保护 usedCodes 的互斥锁
	enableCheckDuplicate = true        // 是否启用错误码重复检查
)

// 预定义的错误实例
var (
	// 系统错误
	ErrInternal = New(1001, "internal server error")
	ErrDatabase = New(1002, "database error")
	ErrCache    = New(1003, "cache error")
	ErrNetwork  = New(1004, "network error")
	ErrTimeout  = New(1005, "request timeout")
	ErrConfig   = New(1006, "invalid configuration")
	ErrResource = New(1007, "resource exhausted")
	ErrService  = New(1008, "service unavailable")
	ErrUnknown  = New(1099, "unknown error")

	// 参数错误
	ErrInvalidParam = New(1100, "invalid parameter")
	ErrMissingParam = New(1101, "missing parameter")
	ErrParamFormat  = New(1102, "parameter format error")
	ErrParamRange   = New(1103, "parameter out of range")
	ErrParamValue   = New(1104, "invalid parameter value")

	// 业务错误
	ErrNotFound         = New(1200, "resource not found")
	ErrAlreadyExists    = New(1201, "resource already exists")
	ErrPermissionDenied = New(1202, "permission denied")
	ErrForbidden        = New(1203, "forbidden access")
	ErrUnauthorized     = New(1204, "unauthorized")
	ErrAccessDenied     = New(1205, "access denied")
	ErrOperation        = New(1206, "operation failed")
	ErrBusinessRule     = New(1207, "business rule violation")

	// 数据错误
	ErrDataValidation  = New(1300, "data validation error")
	ErrDataFormat      = New(1301, "data format error")
	ErrDataCorrupted   = New(1302, "data corrupted")
	ErrDataConsistency = New(1303, "data consistency error")
)

// CodeError 是一个带有错误码的错误类型
// 重命名为 CodeError 以避免与标准库的 error 接口冲突
type CodeError struct {
	code  int
	msg   string
	cause error
}

// Error 实现 error 接口
func (e *CodeError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("code=%d, msg=%s, cause=%v", e.code, e.msg, e.cause)
	}
	return fmt.Sprintf("code=%d, msg=%s", e.code, e.msg)
}

// Code 返回错误码
func (e *CodeError) Code() int {
	return e.code
}

// Unwrap 实现标准库的错误解包接口
func (e *CodeError) Unwrap() error {
	return e.cause
}

// checkDuplicateCode 检查错误码是否已被使用
func checkDuplicateCode(code int) error {
	if !enableCheckDuplicate {
		return nil
	}

	codesMutex.RLock()
	_, exists := usedCodes[code]
	codesMutex.RUnlock()

	if exists {
		return fmt.Errorf("error code %d is already in use", code)
	}

	// 注册新的错误码
	codesMutex.Lock()
	usedCodes[code] = true
	codesMutex.Unlock()

	return nil
}

// New 创建一个新的错误
func New(code int, msg string) error {
	// 确保错误码不小于 1000
	if code < 1000 {
		code = 1000
	}

	// 检查错误码是否重复
	if err := checkDuplicateCode(code); err != nil {
		panic(fmt.Sprintf("duplicate error code: %v", err))
	}

	return &CodeError{
		code: code,
		msg:  msg,
	}
}

// Newf 使用格式化字符串创建一个新的错误
func Newf(code int, format string, args ...interface{}) error {
	// 确保错误码不小于 1000
	if code < 1000 {
		code = 1000
	}

	// 检查错误码是否重复
	if err := checkDuplicateCode(code); err != nil {
		panic(fmt.Sprintf("duplicate error code: %v", err))
	}

	return &CodeError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
}

// Wrap 包装一个错误，添加错误码和消息
func Wrap(err error, code int, msg string) error {
	if err == nil {
		return nil
	}
	// 确保错误码不小于 1000
	if code < 1000 {
		code = 1000
	}

	// 检查错误码是否重复
	if err = checkDuplicateCode(code); err != nil {
		panic(fmt.Sprintf("duplicate error code: %v", err))
	}

	return &CodeError{
		code:  code,
		msg:   msg,
		cause: err,
	}
}

// Wrapf 使用格式化字符串包装一个错误，添加错误码
func Wrapf(err error, code int, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	// 确保错误码不小于 1000
	if code < 1000 {
		code = 1000
	}

	// 检查错误码是否重复
	if err := checkDuplicateCode(code); err != nil {
		panic(fmt.Sprintf("duplicate error code: %v", err))
	}

	return &CodeError{
		code:  code,
		msg:   fmt.Sprintf(format, args...),
		cause: err,
	}
}

// Cause 获取最底层的错误原因
func Cause(err error) error {
	for err != nil {
		cause, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		unwrapped := cause.Unwrap()
		if unwrapped == nil {
			break
		}
		err = unwrapped
	}
	return err
}

// Code 从错误中提取错误码，如果不是 *CodeError 类型，则返回 0
func Code(err error) int {
	if err == nil {
		return 0
	}
	e, ok := err.(*CodeError)
	if !ok {
		// 尝试递归查找包装的错误
		if cause, ok := err.(interface{ Unwrap() error }); ok {
			return Code(cause.Unwrap())
		}
		return 0
	}
	return e.code
}

// IsErrorCode 检查错误是否包含指定的错误码
func IsErrorCode(err error, code int) bool {
	return Code(err) == code
}

// EnableDuplicateCheck 启用或禁用错误码重复检查
func EnableDuplicateCheck(enable bool) {
	enableCheckDuplicate = enable
}
