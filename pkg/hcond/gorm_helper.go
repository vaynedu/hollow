package hcond

import "gorm.io/gorm"

// Conditioner 实现该接口的对象可以被 BuildWhereClause 使用
type Conditioner interface {
	ToSQL() (string, []interface{}, error)
}

// BuildWhereClause 根据条件对象构建 GORM 的 Where 子句
// 若 cond 解析失败,返回原 db 与 error,调用方自己决定是否中断
func BuildWhereClause(db *gorm.DB, cond Conditioner) (*gorm.DB, error) {
	sql, args, err := cond.ToSQL()
	if err != nil {
		return db, err
	}
	if sql == "" {
		return db, nil
	}
	return db.Where(sql, args...), nil
}
