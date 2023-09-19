package orm

// 定义枚举, string 的衍生类型
type op string

const (
	opEq  op = "="
	opLT  op = "<"
	opGT  op = ">"
	opOr  op = "OR"
	opNot op = "NOT"
	opAnd op = "AND"
)

func (o op) String() string {
	return string(o)
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

// 这种设计也是可以的, 只不过子查询调用起来不优雅
//func (p Predicate) Eq(column string, arg any) Predicate {
//	return Predicate{
//		Column: column,
//		Arg:    arg,
//		Op:     "=",
//	}
//}

// Not(C("name")).Eq("Tom")
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// C("id").Eq(12).And(C("name").Eq("Tom"))
func (left Predicate) And(right Predicate) Predicate {

	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

// C("id").Eq(12).Or(C("name").Eq("Tom"))
func (left Predicate) Or(right Predicate) Predicate {

	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

func (p Predicate) expr() {}

type value struct {
	val any
}

func (p value) expr() {}
