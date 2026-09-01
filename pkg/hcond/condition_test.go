package hcond

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCondition_ToSQL(t *testing.T) {
	Convey("Condition.ToSQL", t, func() {
		Convey("原子条件 = 操作符", func() {
			cond := Condition{Operator: OpEq, LHS: "column", RHS: "value"}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "column = ?")
			So(args, ShouldResemble, []interface{}{"value"})
		})

		Convey("原子条件 IN 操作符", func() {
			cond := Condition{Operator: OpIn, LHS: "column", RHS: []interface{}{1, 2, 3}}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "column IN (?, ?, ?)")
			So(args, ShouldResemble, []interface{}{1, 2, 3})
		})

		Convey("IN 操作符 RHS 不是切片", func() {
			cond := Condition{Operator: OpIn, LHS: "status", RHS: "active"}
			sql, args, err := cond.ToSQL()
			So(err, ShouldEqual, ErrRHSNotSlice)
			So(sql, ShouldBeEmpty)
			So(args, ShouldBeNil)
		})

		Convey("逻辑 AND", func() {
			cond := Condition{
				Operator: OpAnd,
				Conditions: []Condition{
					{Operator: OpEq, LHS: "col1", RHS: "val1"},
					{Operator: OpEq, LHS: "col2", RHS: "val2"},
				},
			}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "(col1 = ? AND col2 = ?)")
			So(args, ShouldResemble, []interface{}{"val1", "val2"})
		})

		Convey("逻辑 OR", func() {
			cond := Condition{
				Operator: OpOr,
				Conditions: []Condition{
					{Operator: OpEq, LHS: "col1", RHS: "val1"},
					{Operator: OpEq, LHS: "col2", RHS: "val2"},
				},
			}
			sql, args, err := cond.ToSQL()
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "(col1 = ? OR col2 = ?)")
			So(args, ShouldResemble, []interface{}{"val1", "val2"})
		})

		Convey("逻辑节点 Operator 非 AND/OR 返回 ErrUnsupportedOperator", func() {
			cond := Condition{
				Operator: OpEq, // 错填,逻辑节点只接受 OpAnd/OpOr
				Conditions: []Condition{
					{Operator: OpEq, LHS: "col1", RHS: "val1"},
					{Operator: OpEq, LHS: "col2", RHS: "val2"},
				},
			}
			sql, args, err := cond.ToSQL()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unsupported operator")
			So(sql, ShouldBeEmpty)
			So(args, ShouldBeNil)
		})

		Convey("未支持的操作符返回 ErrUnsupportedOperator", func() {
			cond := Condition{Operator: Op("INVALID"), LHS: "column", RHS: "value"}
			_, _, err := cond.ToSQL()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unsupported operator")
		})
	})
}

func TestParse(t *testing.T) {
	Convey("Parse", t, func() {
		Convey("正常条件加 WHERE 前缀", func() {
			cond := Condition{Operator: OpEq, LHS: "id", RHS: 1}
			sql, args, err := Parse(cond)
			So(err, ShouldBeNil)
			So(sql, ShouldEqual, "WHERE id = ?")
			So(args, ShouldResemble, []interface{}{1})
		})

		Convey("解析错误透传", func() {
			cond := Condition{Operator: OpIn, LHS: "x", RHS: "not-slice"}
			_, _, err := Parse(cond)
			So(err, ShouldEqual, ErrRHSNotSlice)
		})
	})
}
