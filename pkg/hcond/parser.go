package hcond

import "fmt"

// Parse 将 Condition 转换为完整 SQL WHERE 子句(带 WHERE 关键字)
// 当 sql 为空时不返回 "WHERE"
func Parse(cond Condition) (string, []interface{}, error) {
	sql, args, err := cond.ToSQL()
	if err != nil {
		return "", nil, err
	}
	if sql == "" {
		return "", args, nil
	}
	return fmt.Sprintf("WHERE %s", sql), args, nil
}
