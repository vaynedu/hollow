package hcond

import (
	"fmt"
)

// Parse 将 Condition 转换为 SQL WHERE 片段
func Parse(cond Condition) (string, []interface{}, error) {
	sql, args, err := cond.ToSQL()
	return fmt.Sprintf("WHERE %s", sql), args, err
}
