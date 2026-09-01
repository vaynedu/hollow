package hcond

import (
	"fmt"
	"strings"
)

var (
	ErrRHSNotSlice         = fmt.Errorf("RHS must be a slice for IN operator")
	ErrUnsupportedOperator = fmt.Errorf("unsupported operator")
)

// Condition 表示一个条件节点:原子条件或子条件组
type Condition struct {
	Operator   Op          `json:"operator"`
	LHS        string      `json:"lhs"`
	RHS        interface{} `json:"rhs"`
	Conditions []Condition `json:"conditions"`
}

// ToSQL 生成 SQL WHERE 片段(不含 WHERE 关键字)和占位参数
func (c *Condition) ToSQL() (string, []interface{}, error) {
	if len(c.Conditions) == 0 {
		return c.toAtomicSQL()
	}

	clauses := make([]string, 0, len(c.Conditions))
	args := make([]interface{}, 0)
	for i := range c.Conditions {
		sub := c.Conditions[i]
		sql, subArgs, err := sub.ToSQL()
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, sql)
		args = append(args, subArgs...)
	}

	op := OpAnd
	if c.Operator == OpOr {
		op = OpOr
	}
	return fmt.Sprintf("(%s)", strings.Join(clauses, " "+string(op)+" ")), args, nil
}

func (c *Condition) toAtomicSQL() (string, []interface{}, error) {
	switch c.Operator {
	case OpEq, OpNotEq, OpGt, OpLt, OpGte, OpLte:
		return fmt.Sprintf("%s %s ?", c.LHS, c.Operator), []interface{}{c.RHS}, nil
	case OpIn:
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
