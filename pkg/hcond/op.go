package hcond

type Op string

const (
	OpEq    Op = "=="
	OpNotEq Op = "!="
	OpGt    Op = ">"
	OpLt    Op = "<"
	OpGte   Op = ">="
	OpLte   Op = "<="

	OpIn Op = "IN"

	OpAnd Op = "&&"
	OpOr  Op = "||"
	OpNot Op = "!"
)
