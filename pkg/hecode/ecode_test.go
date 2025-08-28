package hecode

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUseCase(t *testing.T) {
	ErrTest001 := New(1001, "test error")
	ErrTest002 := New(1002, "test error")

	Convey("UseCase", t, func() {
		So(ErrTest001, ShouldNotBeNil)
		So(ErrTest002, ShouldNotBeNil)
	})
}

func TestErrorCode(t *testing.T) {
	Convey("ErrorCode", t, func() {
		code := 1001
		msg := "test error"
		err := New(code, msg)
		So(err, ShouldNotBeNil)
		So(Code(err), ShouldEqual, code)
		So(IsErrorCode(err, code), ShouldBeTrue)
	})
}

func TestErrorCodeWrap(t *testing.T) {
	Convey("ErrorCodeWrap", t, func() {
		code := 1001
		msg := "test error"
		err := New(code, msg)
		So(err, ShouldNotBeNil)
		So(Code(err), ShouldEqual, code)
		So(IsErrorCode(err, code), ShouldBeTrue)
	})
}

// 测试错误码小于 1000 的情况
func TestErrorCodeBelow1000(t *testing.T) {
	Convey("ErrorCodeBelow1000", t, func() {
		// 测试错误码小于 1000 的情况，应该被修正为 1000
		code := 500
		err := New(code, "test error")
		So(Code(err), ShouldEqual, 1000) // 应该被修正为 1000
		So(IsErrorCode(err, 1000), ShouldBeTrue)
		
		// 测试 Newf 函数
		err = Newf(code, "test error %d", 1)
		So(Code(err), ShouldEqual, 1000)
		
		// 测试 Wrap 函数
		origErr := errors.New("original error")
		err = Wrap(origErr, code, "wrapped error")
		So(Code(err), ShouldEqual, 1000)
		
		// 测试 Wrapf 函数
		err = Wrapf(origErr, code, "wrapped error %d", 1)
		So(Code(err), ShouldEqual, 1000)
	})
}

// 测试错误包装和解包功能
func TestErrorWrappingAndUnwrapping(t *testing.T) {
	Convey("ErrorWrappingAndUnwrapping", t, func() {
		// 创建原始错误
		origErr := errors.New("original error")
		
		// 包装错误
		code := 1001
		wrappedErr := Wrap(origErr, code, "wrapped error")
		
		// 检查包装后的错误码
		So(Code(wrappedErr), ShouldEqual, code)
		So(IsErrorCode(wrappedErr, code), ShouldBeTrue)
		
		// 检查错误解包
		So(errors.Is(wrappedErr, origErr), ShouldBeTrue)
		unwrappedErr := Cause(wrappedErr)
		So(unwrappedErr, ShouldEqual, origErr)
		
		// 测试多层包装
		code2 := 1002
		wrappedErr2 := Wrap(wrappedErr, code2, "second wrapped error")
		So(Code(wrappedErr2), ShouldEqual, code2)
		So(errors.Is(wrappedErr2, wrappedErr), ShouldBeTrue)
		So(errors.Is(wrappedErr2, origErr), ShouldBeTrue)
		unwrappedErr2 := Cause(wrappedErr2)
		So(unwrappedErr2, ShouldEqual, origErr)
	})
}

// 测试错误码重复检查功能
func TestDuplicateErrorCodeCheck(t *testing.T) {
	Convey("DuplicateErrorCodeCheck", t, func() {
		// 保存原始的启用状态
		originalEnableState := enableCheckDuplicate
		// 测试完成后恢复原始状态
		defer func() { enableCheckDuplicate = originalEnableState }()
		
		// 测试重复的错误码是否会触发panic
		// 先创建一个临时的唯一错误码
		uniqueCode := 2000
		// 保存当前usedCodes状态
		currentUsedCodes := make(map[int]bool)
		codesMutex.RLock()
		for code := range usedCodes {
			currentUsedCodes[code] = true
		}
		codesMutex.RUnlock()
		
		// 确保uniqueCode是未使用的
		for currentUsedCodes[uniqueCode] {
			uniqueCode++
		}
		
		// 禁用错误码检查，创建测试用的错误
		EnableDuplicateCheck(false)
		testErr := New(uniqueCode, "test error for duplicate check")
		EnableDuplicateCheck(true)
		
		// 确保我们能成功创建一个错误
		So(testErr, ShouldNotBeNil)
		So(Code(testErr), ShouldEqual, uniqueCode)
		
		// 注册这个错误码到usedCodes
		codesMutex.Lock()
		usedCodes[uniqueCode] = true
		codesMutex.Unlock()
		
		// 测试再次使用同一个错误码是否会触发panic
		Convey("Using duplicate error code should panic", func() {
			success := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						success = true
					}
				}()
				// 尝试创建一个重复的错误码
				New(uniqueCode, "duplicate error code")
			}()
			// 验证是否触发了panic
			So(success, ShouldBeTrue)
		})
		
		// 测试禁用重复检查后，是否可以使用重复的错误码
		Convey("After disabling duplicate check, duplicate code should be allowed", func() {
			// 禁用重复检查
			EnableDuplicateCheck(false)
			// 尝试创建一个重复的错误码，这次不应该panic
			var success bool
			func() {
				defer func() {
					if r := recover(); r == nil {
						success = true
					}
				}()
				err := New(uniqueCode, "duplicate error code after disabling check")
				So(err, ShouldNotBeNil)
				So(Code(err), ShouldEqual, uniqueCode)
			}()
			// 验证是否没有触发panic
			So(success, ShouldBeTrue)
		})
		
		// 清理测试数据
		codesMutex.Lock()
		delete(usedCodes, uniqueCode)
		codesMutex.Unlock()
	})
}

// 测试EnableDuplicateCheck函数功能
func TestEnableDuplicateCheck(t *testing.T) {
	Convey("EnableDuplicateCheck", t, func() {
		// 保存原始的启用状态
		originalEnableState := enableCheckDuplicate
		// 测试完成后恢复原始状态
		defer func() { enableCheckDuplicate = originalEnableState }()
		
		// 测试启用和禁用功能
		EnableDuplicateCheck(false)
		So(enableCheckDuplicate, ShouldBeFalse)
		
		EnableDuplicateCheck(true)
		So(enableCheckDuplicate, ShouldBeTrue)
	})
}