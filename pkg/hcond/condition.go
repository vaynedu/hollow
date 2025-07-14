package hcond

import (
	"fmt"
	"strings"
)

// 包级错误变量定义
var (
	ErrRHSNotSlice = fmt.Errorf("RHS must be a slice for IN operator")
	ErrUnsupportedOperator = fmt.Errorf("unsupported operator")
)

// Condition 表示一个条件节点：可以是原子条件，也可以是一个子条件组
 type Condition struct {
	Operator   string      `json:"operator"`   // 支持 =, !=, >, <, IN, &&, || 等
	LHS        string      `json:"lhs"`        // 左值字段名（仅在原子条件中使用）
	RHS        interface{} `json:"rhs"`        // 右值（仅在原子条件中使用）
	Conditions []Condition `json:"conditions"` // 子条件（仅在逻辑条件中使用）
}

func (c *Condition) ToSQL() (string, []interface{}, error) {
	if len(c.Conditions) == 0 {
		// 原子条件
		sql, args, err := c.toAtomicSQL()
		if err != nil {
			return "", nil, err
		}
		return sql, args, nil
	}

	// 逻辑条件
	var clauses []string
	var args []interface{}

	for _, sub := range c.Conditions {
		sql, subArgs, err := sub.ToSQL()
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, sql)
		args = append(args, subArgs...)
	}

	op := "AND"
	if c.Operator == "||" {
		op = "OR"
	}

	return fmt.Sprintf("(%s)", strings.Join(clauses, " "+op+" ")), args, nil
}

func (c *Condition) toAtomicSQL() (string, []interface{}, error) {
	const (
		Equal        = "="
		NotEqual     = "!="
		GreaterThan  = ">"
		LessThan     = "<"
		GreaterEqual = ">="
		LessEqual    = "<="
		In           = "IN"
	)

	switch c.Operator {
	case Equal, NotEqual, GreaterThan, LessThan, GreaterEqual, LessEqual:
		return fmt.Sprintf("%s %s ?", c.LHS, c.Operator), []interface{}{c.RHS}, nil
	case In:
		values, ok := c.RHS.([]interface{})
		if !ok {
			return "", nil, ErrRHSNotSlice
		}
		placeholders := make([]string, len(values))
		for i := range values {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s IN (%s)", c.LHS, strings.Join(placeholders, ", ")), values, nil
	default:
		return "", nil, fmt.Errorf("%w: %s", ErrUnsupportedOperator, c.Operator)
	}
}

