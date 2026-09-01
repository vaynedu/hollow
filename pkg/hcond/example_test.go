package hcond

import "fmt"

func ExampleCondition_ToSQL() {
	cond := Condition{
		Operator: OpAnd,
		Conditions: []Condition{
			{Operator: OpEq, LHS: "status", RHS: "active"},
			{Operator: OpGt, LHS: "age", RHS: 18},
		},
	}
	sql, args, _ := cond.ToSQL()
	fmt.Println(sql)
	fmt.Println(args)
	// Output:
	// (status = ? AND age > ?)
	// [active 18]
}
