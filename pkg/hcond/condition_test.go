package hcond

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestCondition_ToSQL(t *testing.T) {
	// 测试错误处理的用例
	var tests = []struct {
		name         string
		condition    Condition
		expectedSQL  string
		expectedArgs []interface{}
	}{{
		name: "原子条件 = 操作符",
		condition: Condition{
			Operator: "=",
			LHS:      "column",
			RHS:      "value",
		},
		expectedSQL:  "column = ?",
		expectedArgs: []interface{}{"value"},
	}, {
		name: "原子条件 IN 操作符",
		condition: Condition{
			Operator: "IN",
			LHS:      "column",
			RHS:      []interface{}{1, 2, 3},
		},
		expectedSQL:  "column IN (?, ?, ?)",
		expectedArgs: []interface{}{1, 2, 3},
	}, {
		name: "逻辑条件 && 操作符",
		condition: Condition{
			Operator: "&&",
			Conditions: []Condition{
				{
					Operator: "=",
					LHS:      "col1",
					RHS:      "val1",
				},
				{
					Operator: "=",
					LHS:      "col2",
					RHS:      "val2",
				},
			},
		},
		expectedSQL:  "(col1 = ? && col2 = ?)",
		expectedArgs: []interface{}{"val1", "val2"},
	}, {
		name: "逻辑条件 || 操作符",
		condition: Condition{
			Operator: "||",
			Conditions: []Condition{
				{
					Operator: "=",
					LHS:      "col1",
					RHS:      "val1",
				},
				{
					Operator: "=",
					LHS:      "col2",
					RHS:      "val2",
				},
			},
		},
		expectedSQL:  "(col1 = ? OR col2 = ?)",
		expectedArgs: []interface{}{"val1", "val2"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualSQL, actualArgs, actualErr := tt.condition.ToSQL()
			if actualErr == nil && actualSQL != tt.expectedSQL {
				t.Errorf("期望 SQL: %s, 实际 SQL: %s", tt.expectedSQL, actualSQL)
			}
			if actualErr == nil && !reflect.DeepEqual(actualArgs, tt.expectedArgs) {
				t.Errorf("期望参数: %v, 实际参数: %v", tt.expectedArgs, actualArgs)
			}
		})
	}

}

func TestCondition_toAtomicSQL(t *testing.T) {
	// 测试错误处理的用例
	condInvalidOp := Condition{
		Operator: "INVALID",
		LHS:      "column",
		RHS:      "value",
	}
	expectedErrInvalidOp := fmt.Errorf("unsupported operator: INVALID")

	actualSQL, actualArgs, actualErr := condInvalidOp.toAtomicSQL()
	if actualErr == nil || errors.Is(actualErr, expectedErrInvalidOp) {
		t.Errorf("期望错误: %v, 实际错误: %v", expectedErrInvalidOp, actualErr)
	}

	condInvalidIn := Condition{
		Operator: "IN",
		LHS:      "column",
		RHS:      "not a slice",
	}
	expectedErrInvalidIn := fmt.Errorf("RHS must be a slice for IN operator")

	actualSQL, actualArgs, actualErr = condInvalidIn.toAtomicSQL()
	if actualErr == nil || errors.Is(actualErr, expectedErrInvalidIn) {
		t.Errorf("期望错误: %v, 实际错误: %v", expectedErrInvalidIn, actualErr)
	}

	// 原有的测试用例
	// 可以添加更多的测试用例来覆盖不同的原子条件操作符
	// 这里简单示例一个测试用例
	cond := Condition{
		Operator: "=",
		LHS:      "column",
		RHS:      "value",
	}
	expectedSQL := "column = ?"
	expectedArgs := []interface{}{"value"}

	actualSQL, actualArgs, actualErr = cond.toAtomicSQL()
	if actualSQL != expectedSQL {
		t.Errorf("期望 SQL: %s, 实际 SQL: %s", expectedSQL, actualSQL)
	}
	if !reflect.DeepEqual(actualArgs, expectedArgs) {
		t.Errorf("期望参数: %v, 实际参数: %v", expectedArgs, actualArgs)
	}
}
