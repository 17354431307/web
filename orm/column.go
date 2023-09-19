package orm

type Column struct {
	name  string
	alias string
}

func C(name string) Column {
	return Column{name: name}
}

// 这个设计是不可变的设计 immutable, 和 react 的理念一样
func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: c.valueOf(arg),
	}
}

func (c Column) valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{
			val: arg,
		}
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: value{val: arg},
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: value{val: arg},
	}
}

func (c Column) expr() {}

func (c Column) selectable() {

}
