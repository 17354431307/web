package orm

// Expression 是一个标记接口，代表表达式
type Expression interface {
	expr()
}

// RawExpr 代表着原生表达式
// Raw 不是 Row，不要写错了
type RawExpr struct {
	raw  string
	args []any
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

func (r RawExpr) selectable() {}

func (r RawExpr) expr() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
