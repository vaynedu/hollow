package hcond

import (
	"gorm.io/gorm"
)

// Conditioner 定义一个接口，确保传入的条件对象实现了 ToSQL 方法
type Conditioner interface {
	ToSQL() (string, []interface{})
}

// BuildWhereClause 根据条件对象构建 GORM 的 Where 子句
func BuildWhereClause(db *gorm.DB, cond Conditioner) *gorm.DB {
	sql, args := cond.ToSQL()
	if sql == "" {
		return db
	}
	return db.Where(sql, args...)
}
