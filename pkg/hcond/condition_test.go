package hcond

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
	}, {name: "原子条件 IN 操作符",
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
	if actualSQL != "" {
		t.Errorf("期望 SQL: %s, 实际 SQL: %s", "", actualSQL)
	}
	if actualArgs != nil {
		t.Errorf("期望参数: %v, 实际参数: %v", nil, actualArgs)
	}

	// 测试IN操作符RHS非切片的错误情况
	condInvalidIn := Condition{
		Operator: "IN",
		LHS:      "status",
		RHS:      "active", // 非切片类型
	}

	// 使用包级错误变量进行比较
	actualSQL, actualArgs, actualErr = condInvalidIn.toAtomicSQL()
	assert.ErrorIs(t, actualErr, ErrRHSNotSlice)
	assert.Empty(t, actualSQL)
	assert.Nil(t, actualArgs)
}