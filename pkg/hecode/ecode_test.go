package hecode

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUseCase(t *testing.T) {
	ErrTest001 := New(1000001, "test error")
	ErrTest002 := New(1000002, "test error")

	Convey("UseCase", t, func() {
		So(ErrTest001, ShouldNotBeNil)
		So(ErrTest002, ShouldNotBeNil)
	})
}

func TestErrorCode(t *testing.T) {
	Convey("ErrorCode", t, func() {
		code := 1000003
		msg := "test error"
		err := New(code, msg)
		So(err, ShouldNotBeNil)
		So(Code(err), ShouldEqual, code)
		So(IsErrorCode(err, code), ShouldBeTrue)
	})
}

func TestErrorCodeWrap(t *testing.T) {
	Convey("ErrorCodeWrap", t, func() {
		code := 1000004
		msg := "test error"
		err := New(code, msg)
		So(err, ShouldNotBeNil)
		So(Code(err), ShouldEqual, code)
		So(IsErrorCode(err, code), ShouldBeTrue)
	})
}

// 测试错误创建功能 (New 和 Newf)
func TestErrorNew(t *testing.T) {
	Convey("ErrorNew", t, func() {
		// 测试基本的 New 函数
		code := 1000005
		msg := "test error"
		err := New(code, msg)
		So(err, ShouldNotBeNil)
		So(Code(err), ShouldEqual, code)
		So(err.Error(), ShouldContainSubstring, msg)

		// 测试 Newf 函数
		code2 := 1000006
		format := "test error %d"
		arg := 2
		err2 := Newf(code2, format, arg)
		So(err2, ShouldNotBeNil)
		So(Code(err2), ShouldEqual, code2)
		So(err2.Error(), ShouldContainSubstring, fmt.Sprintf(format, arg))
	})
}

// 测试错误包装功能 (Wrap, Wrapf, WrapError)
func TestErrorWrap(t *testing.T) {
	Convey("ErrorWrap", t, func() {
		// 创建基础错误
		baseCode := 1000007
		baseErr := New(baseCode, "base error")

		// 测试 Wrap 函数
		wrappedErr := Wrap(baseErr, "wrapped message")
		So(wrappedErr, ShouldNotBeNil)
		So(Code(wrappedErr), ShouldEqual, baseCode) // 保持原始错误码
		So(wrappedErr.Error(), ShouldContainSubstring, "wrapped message")
		So(IsErrorCode(wrappedErr, baseCode), ShouldBeTrue)

		// 检查 wrappedErr 的 cause 是否为 baseErr
		e, ok := wrappedErr.(*EcodeError)
		So(ok, ShouldBeTrue)
		So(e.cause, ShouldEqual, baseErr)

		// 测试 Wrapf 函数
		format := "wrapped message with arg: %d"
		arg := 1
		wrappedErr2 := Wrapf(baseErr, format, arg)
		So(wrappedErr2, ShouldNotBeNil)
		So(Code(wrappedErr2), ShouldEqual, baseCode)
		So(wrappedErr2.Error(), ShouldContainSubstring, fmt.Sprintf(format, arg))

		// 测试 WrapError 函数
		errorErr := errors.New("original error")
		wrappedErr3 := WrapError(baseErr, errorErr)
		So(wrappedErr3, ShouldNotBeNil)
		So(Code(wrappedErr3), ShouldEqual, baseCode)
		So(wrappedErr3.Error(), ShouldContainSubstring, errorErr.Error())

		// 测试 Wrap nil 错误
		nilWrapped := Wrap(nil, "should be nil")
		So(nilWrapped, ShouldBeNil)

		// 测试 Wrapf nil 错误
		nilWrapped2 := Wrapf(nil, "should be %s", "nil")
		So(nilWrapped2, ShouldBeNil)

		// 测试包装非 EcodeError 类型的错误
		stdErr := errors.New("standard error")
		wrappedStdErr := Wrap(stdErr, "wrapped standard error")
		So(wrappedStdErr, ShouldNotBeNil)
		So(Code(wrappedStdErr), ShouldEqual, ErrCodeUnknown) // 应该使用未知错误码
		So(wrappedStdErr.Error(), ShouldContainSubstring, "wrapped standard error")

		// 检查 wrappedStdErr 的 cause 是否为 stdErr
		e2, ok := wrappedStdErr.(*EcodeError)
		So(ok, ShouldBeTrue)
		So(e2.cause, ShouldEqual, stdErr)

		// 测试 Wrapf 非 EcodeError 类型的错误
		wrappedStdErr2 := Wrapf(stdErr, "wrapped standard error: %s", "test")
		So(wrappedStdErr2, ShouldNotBeNil)
		So(Code(wrappedStdErr2), ShouldEqual, ErrCodeUnknown)
		So(wrappedStdErr2.Error(), ShouldContainSubstring, "wrapped standard error")
		So(wrappedStdErr2.Error(), ShouldContainSubstring, "test")
	})
}

// 测试错误原因功能 (Cause, GetCause)
func TestErrorCause(t *testing.T) {
	Convey("ErrorCause", t, func() {
		// 创建嵌套错误
		origErr := errors.New("original error")
		code := 1000010
		wrappedErr1 := NewErrorWithCause(code, "first wrapped", origErr)
		wrappedErr2 := Wrap(wrappedErr1, "second wrapped")

		// 测试 Cause 函数
		rootCause := Cause(wrappedErr2)
		So(rootCause, ShouldEqual, origErr)

		// 测试 GetCause 方法
		e, ok := wrappedErr2.(*EcodeError)
		So(ok, ShouldBeTrue)
		rootCause2 := e.GetCause()
		So(rootCause2, ShouldEqual, origErr)

		// 测试非嵌套错误的 Cause
		baseErr := New(1000011, "base error")
		baseCause := Cause(baseErr)
		So(baseCause, ShouldEqual, baseErr) // 没有嵌套错误时，Cause 应该返回自身

		// 测试标准错误的 Cause
		stdErr := errors.New("standard error")
		stdCause := Cause(stdErr)
		So(stdCause, ShouldEqual, stdErr)

		// 测试多层嵌套错误的 Cause
		err1 := errors.New("level 1 error")
		err2 := Wrap(err1, "level 2 error")
		err3 := Wrap(err2, "level 3 error")
		err4 := Wrap(err3, "level 4 error")
		root := Cause(err4)
		So(root, ShouldEqual, err1)
	})
}

// 测试新添加的辅助函数
func TestErrorHelperFunctions(t *testing.T) {
	Convey("ErrorHelperFunctions", t, func() {
		// 测试 NewErrorWithCause 函数
		code := 1000012
		msg := "error with cause"
		origErr := errors.New("original cause")
		err := NewErrorWithCause(code, msg, origErr)
		So(err, ShouldNotBeNil)
		So(Code(err), ShouldEqual, code)
		So(err.Error(), ShouldContainSubstring, msg)
		So(Cause(err), ShouldEqual, origErr)

		// 测试 GetErrorCodeMessage 函数
		unknownMsg := GetErrorCodeMessage(ErrCodeUnknown)
		So(unknownMsg, ShouldEqual, "unknown error")

		customMsg := GetErrorCodeMessage(2000) // 自定义错误码
		So(customMsg, ShouldContainSubstring, "2000")

		// 测试 WithMessage 函数
		baseErr := New(1000013, "base error")
		newMsgErr := WithMessage(baseErr, "new message")
		So(newMsgErr, ShouldNotBeNil)
		So(Code(newMsgErr), ShouldEqual, 1000013) // 保持原始错误码
		So(newMsgErr.Error(), ShouldContainSubstring, "new message")

		// 测试 WithMessagef 函数
		format := "formatted message: %s"
		arg := "test"
		formattedMsgErr := WithMessagef(baseErr, format, arg)
		So(formattedMsgErr, ShouldNotBeNil)
		So(Code(formattedMsgErr), ShouldEqual, 1000013)
		So(formattedMsgErr.Error(), ShouldContainSubstring, fmt.Sprintf(format, arg))

		// 测试 WithMessage 处理非 EcodeError 类型的错误
		stdErr := errors.New("standard error")
		stdWithMsg := WithMessage(stdErr, "enhanced message")
		So(stdWithMsg, ShouldNotBeNil)
		So(Code(stdWithMsg), ShouldEqual, ErrCodeUnknown)
		So(stdWithMsg.Error(), ShouldContainSubstring, "enhanced message")

		// 检查 stdWithMsg 的 cause 是否为 stdErr
		e, ok := stdWithMsg.(*EcodeError)
		So(ok, ShouldBeTrue)
		So(e.cause, ShouldEqual, stdErr)
	})
}

// 测试预定义错误常量
func TestPredefinedErrors(t *testing.T) {
	Convey("PredefinedErrors", t, func() {
		// 验证系统错误
		So(ErrInternal, ShouldNotBeNil)
		So(Code(ErrInternal), ShouldEqual, 1001)
		So(ErrInternal.Error(), ShouldContainSubstring, "internal server error")

		// 验证参数错误
		So(ErrInvalidParam, ShouldNotBeNil)
		So(Code(ErrInvalidParam), ShouldEqual, 1100)
		So(ErrInvalidParam.Error(), ShouldContainSubstring, "invalid parameter")

		// 验证业务错误
		So(ErrNotFound, ShouldNotBeNil)
		So(Code(ErrNotFound), ShouldEqual, 1200)
		So(ErrNotFound.Error(), ShouldContainSubstring, "resource not found")

		// 验证数据错误
		So(ErrDataValidation, ShouldNotBeNil)
		So(Code(ErrDataValidation), ShouldEqual, 1300)
		So(ErrDataValidation.Error(), ShouldContainSubstring, "data validation error")

		// 验证数据库错误
		So(ErrDatabase, ShouldNotBeNil)
		So(Code(ErrDatabase), ShouldEqual, 1400)
		So(ErrDatabase.Error(), ShouldContainSubstring, "database error")

		// 验证Redis错误
		So(ErrRedisConnection, ShouldNotBeNil)
		So(Code(ErrRedisConnection), ShouldEqual, 1500)
		So(ErrRedisConnection.Error(), ShouldContainSubstring, "redis connection error")
	})
}

// 测试接口方法
func TestInterfaceMethods(t *testing.T) {
	Convey("InterfaceMethods", t, func() {
		err := New(1000014, "test error")
		e, ok := err.(*EcodeError)
		So(ok, ShouldBeTrue)

		// 测试 Code 方法
		So(e.Code(), ShouldEqual, 1000014)

		// 测试 GetMessage 方法
		So(e.GetMessage(), ShouldEqual, "test error")

		// 测试 GetCause 方法 - 没有嵌套错误时
		So(e.GetCause(), ShouldBeNil) // 没有嵌套错误时，GetCause 应该返回 nil

		// 测试单层嵌套错误
		wrappedErr := Wrap(e, "wrapped error")
		e2, ok := wrappedErr.(*EcodeError)
		So(ok, ShouldBeTrue)
		// GetCause 会递归查找最底层的错误原因，所以这里应该返回原始错误 e
		So(e2.GetCause(), ShouldEqual, e)

		// 测试多层嵌套错误
		wrappedErr2 := Wrap(wrappedErr, "double wrapped error")
		e3, ok := wrappedErr2.(*EcodeError)
		So(ok, ShouldBeTrue)
		// 即使是多层嵌套，GetCause 也应该返回最底层的原始错误 e
		So(e3.GetCause(), ShouldEqual, e)
	})
}

// 测试边界情况
func TestEdgeCases(t *testing.T) {
	Convey("EdgeCases", t, func() {
		// 测试错误码刚好等于 1000 的情况
		code := 1000
		err := New(code, "error code exactly 1000")
		So(Code(err), ShouldEqual, code)

		// 测试空字符串消息
		err = New(1000015, "")
		So(err, ShouldNotBeNil)
		So(Code(err), ShouldEqual, 1000015)
		So(err.Error(), ShouldContainSubstring, "code=1000015")

		// 测试包装自身
		err = New(1000016, "base error")
		wrappedSelf := Wrap(err, "wrapped self")
		So(wrappedSelf, ShouldNotBeNil)
		So(Code(wrappedSelf), ShouldEqual, 1000016)
		So(wrappedSelf.Error(), ShouldContainSubstring, "wrapped self")

		// 测试错误链很长的情况
		var lastErr error = New(1000017, "level 1")
		for i := 2; i <= 10; i++ {
			lastErr = Wrap(lastErr, fmt.Sprintf("level %d", i))
		}
		So(lastErr, ShouldNotBeNil)
		rootCause := Cause(lastErr)
		So(rootCause.Error(), ShouldContainSubstring, "level 1")
	})
}
