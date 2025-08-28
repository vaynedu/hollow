package hecode

import (
	"fmt"
	"sync"
)

// 全局变量，用于跟踪已使用的错误码
var (
	usedCodes            = make(map[int]bool) // 存储已使用的错误码
	codesMutex           = &sync.RWMutex{}    // 用于保护 usedCodes 的互斥锁
	enableCheckDuplicate = true               // 是否启用错误码重复检查
)

const (
	// 未知错误
	ErrCodeUnknown = 1099
)

// 预定义的错误实例
var (
	// 系统错误
	ErrInternal = New(1001, "internal server error")
	ErrCache    = New(1002, "cache error")
	ErrNetwork  = New(1003, "network error")
	ErrTimeout  = New(1004, "request timeout")
	ErrConfig   = New(1005, "invalid configuration")
	ErrResource = New(1006, "resource exhausted")
	ErrService  = New(1007, "service unavailable")
	ErrUnknown  = New(ErrCodeUnknown, "unknown error")

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

	// 数据库错误
	ErrDatabase               = New(1400, "database error")
	ErrDBConnection           = New(1401, "database connection error")
	ErrDBQuery                = New(1402, "database query error")
	ErrDBTransaction          = New(1403, "database transaction error")
	ErrDBRollback             = New(1404, "database rollback error")
	ErrDBInsert               = New(1405, "database insert error")
	ErrDBUpdate               = New(1406, "database update error")
	ErrDBDelete               = New(1407, "database delete error")
	ErrDBForeignKey           = New(1408, "database foreign key error")
	ErrDBUnique               = New(1409, "database unique constraint error")
	ErrDBIndex                = New(1410, "database index error")
	ErrDBLock                 = New(1411, "database lock error")
	ErrDBTimeout              = New(1412, "database timeout error")
	ErrDBConnectionLimit      = New(1413, "database connection limit error")
	ErrDBTransactionIsolation = New(1414, "database transaction isolation error")
	ErrDBTransactionDeadlock  = New(1415, "database transaction deadlock error")
	ErrDBTransactionRollback  = New(1416, "database transaction rollback error")
	ErrDBTransactionCommit    = New(1417, "database transaction commit error")

	// redis错误
	ErrRedisConnection = New(1500, "redis connection error")
)

// EcodeError 是一个带有错误码的错误类型
type EcodeError struct {
	code  int
	msg   string
	cause error
}

// Error 实现 error 接口
func (e *EcodeError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("code=%d, msg=%s, cause=%v", e.code, e.msg, e.cause)
	}
	return fmt.Sprintf("code=%d, msg=%s", e.code, e.msg)
}

// Code 返回错误码
func (e *EcodeError) Code() int {
	return e.code
}

// Unwrap 实现标准库的错误解包接口
func (e *EcodeError) Unwrap() error {
	return e.cause
}

// GetCause 获取错误的根本原因
// 注意：该方法会递归查找最底层的错误原因，而不仅仅是直接原因
// 通过调用Cause函数对e.cause进行递归解包，获取最底层的错误原因
func (e *EcodeError) GetCause() error {
	return Cause(e.cause)
}

// GetMessage 获取错误消息
func (e *EcodeError) GetMessage() string {
	return e.msg
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
		panic("error code must be at least 1000") // 暴力panic
	}

	// 检查错误码是否重复
	if err := checkDuplicateCode(code); err != nil {
		panic(fmt.Sprintf("duplicate error code: %v", err))
	}

	return &EcodeError{
		code: code,
		msg:  msg,
		// 当使用 New 创建错误时，cause 默认为 nil
		// 因为这是一个新创建的错误，没有原始错误可以包装
		cause: nil,
	}
}

// Newf 使用格式化字符串创建一个新的错误
func Newf(code int, format string, args ...interface{}) error {
	// 确保错误码不小于 1000
	if code < 1000 {
		panic("error code must be at least 1000") // 暴力panic
	}

	// 检查错误码是否重复
	if err := checkDuplicateCode(code); err != nil {
		panic(fmt.Sprintf("duplicate error code: %v", err))
	}

	return &EcodeError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
		// 当使用 Newf 创建错误时，cause 默认为 nil
		cause: nil,
	}
}

// Wrap 包装一个错误，添加错误码和消息
// cause 用于保存原始错误，形成错误链，方便追踪完整的错误上下文
func Wrap(err error, msg string) error {
	if err == nil {
		return nil // 如果原始错误为 nil，则返回 nil，不需要创建新的错误对象
	}

	// 先尝试将错误转换为 *EcodeError
	if ec, ok := err.(*EcodeError); ok {
		// 如果是 EcodeError 类型，保留原始错误码，包装原始错误作为 cause
		return &EcodeError{
			code:  ec.code, // 使用原始错误的错误码
			msg:   msg,     // 使用新的错误消息
			cause: err,     // 保留原始错误作为 cause
		}
	}

	// 如果不是 EcodeError 类型，使用默认的未知错误码，包装原始错误作为 cause
	return &EcodeError{
		code:  ErrCodeUnknown, // 使用未知错误码
		msg:   msg,            // 使用新的错误消息
		cause: err,            // 保留原始错误作为 cause
	}
}

// Wrapf 使用格式化字符串包装一个错误，添加错误码
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// 先尝试将错误转换为 *EcodeError
	if ec, ok := err.(*EcodeError); ok {
		// 如果是 EcodeError 类型，保留原始错误码，包装原始错误作为 cause
		return &EcodeError{
			code:  ec.code,                      // 使用原始错误的错误码
			msg:   fmt.Sprintf(format, args...), // 使用格式化的新错误消息
			cause: err,                          // 保留原始错误作为 cause
		}
	}

	// 如果不是 EcodeError 类型，使用默认的未知错误码，包装原始错误作为 cause
	return &EcodeError{
		code:  ErrCodeUnknown,               // 使用未知错误码
		msg:   fmt.Sprintf(format, args...), // 使用格式化的新错误消息
		cause: err,                          // 保留原始错误作为 cause
	}
}

// WrapError 包装一个错误，添加错误码和消息
func WrapError(err error, newErr error) error {
	if err == nil {
		return nil
	}

	// 先尝试将错误转换为 *EcodeError
	if ec, ok := err.(*EcodeError); ok {
		// 如果是 EcodeError 类型，保留原始错误码，包装原始错误作为 cause
		return &EcodeError{
			code:  ec.code,
			msg:   newErr.Error(),
			cause: err,
		}
	}

	return &EcodeError{
		code:  ErrCodeUnknown,
		msg:   newErr.Error(),
		cause: err,
	}
}

// Cause 获取最底层的错误原因
func Cause(err error) error {
	// 如果err为nil，直接返回nil
	if err == nil {
		return nil
	}

	// 递归查找最底层的错误原因
	for {
		cause, ok := err.(interface{ Unwrap() error })
		if !ok {
			// 如果不是可解包的错误，返回当前错误
			return err
		}
		unwrapped := cause.Unwrap()
		if unwrapped == nil {
			// 如果解包后为nil，返回当前错误
			return err
		}
		err = unwrapped
	}
}

// Code 从错误中提取错误码，如果不是 *EcodeError 类型，则返回 0
func Code(err error) int {
	if err == nil {
		return 0
	}
	e, ok := err.(*EcodeError)
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

// IsError 检查错误是否与目标错误具有相同的错误码
func IsError(err error, target error) bool {
	if err == nil || target == nil {
		return err == target
	}

	targetCode := Code(target)
	return IsErrorCode(err, targetCode)
}

// NewErrorWithCause 创建一个带有原因的错误
func NewErrorWithCause(code int, msg string, cause error) error {
	// 确保错误码不小于 1000
	if code < 1000 {
		panic("error code must be at least 1000")
	}

	// 检查错误码是否重复
	if err := checkDuplicateCode(code); err != nil {
		panic(fmt.Sprintf("duplicate error code: %v", err))
	}

	return &EcodeError{
		code:  code,
		msg:   msg,
		cause: cause,
	}
}

// GetErrorCodeMessage 根据错误码获取预定义的错误消息
func GetErrorCodeMessage(code int) string {
	// 这里可以根据需要添加更多的错误码到错误消息的映射
	switch code {
	case ErrCodeUnknown:
		return "unknown error"
	default:
		return fmt.Sprintf("error with code %d", code)
	}
}

// WithMessage 返回带有新消息的错误，保持错误码不变
func WithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*EcodeError)
	if !ok {
		return &EcodeError{
			code:  ErrCodeUnknown,
			msg:   msg,
			cause: err,
		}
	}

	return &EcodeError{
		code:  e.code,
		msg:   msg,
		cause: e,
	}
}

// WithMessagef 返回带有格式化新消息的错误，保持错误码不变
func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return WithMessage(err, fmt.Sprintf(format, args...))
}
