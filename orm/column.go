package orm

type Column struct {
	name string
}

func C(name string) Column {
	return Column{name: name}
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
