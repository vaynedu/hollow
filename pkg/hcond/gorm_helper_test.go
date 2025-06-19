package hcond

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockCondition 实现 Conditioner 接口，用于测试
type MockCondition struct {
	SQL  string
	Args []interface{}
}

func (m MockCondition) ToSQL() (string, []interface{}) {
	return m.SQL, m.Args
}

func TestBuildWhereClause(t *testing.T) {
	// 初始化一个模拟的 gorm.DB 实例
	mockDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("初始化 gorm.DB 实例失败: %v", err)
	}

	// 测试用例 1: 空 SQL
	cond1 := MockCondition{SQL: "", Args: nil}
	result1 := BuildWhereClause(mockDB, cond1)
	if result1 != mockDB {
		t.Errorf("期望返回原始的 db 实例，实际返回不同的实例")
	}

	// 测试用例 2: 非空 SQL
	sql := "column = ?"
	args := []interface{}{"value"}
	cond2 := MockCondition{SQL: sql, Args: args}
	result2 := BuildWhereClause(mockDB, cond2)
	if result2 == mockDB {
		t.Errorf("期望返回一个新的 db 实例，实际返回原始实例")
	}
}
