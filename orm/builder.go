package orm

import (
	"github.com/Moty1999/web/orm/internal/errs"
	"strings"
)

type builder struct {
	sb   *strings.Builder
	args []any
}

func (b *builder) buildPredicates(ps []Predicate, m *Model) error {

	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p.And(ps[i])
	}

	return b.buildExpression(p, m)
}

func (b *builder) buildExpression(expr Expression, m *Model) error {

	switch exp := expr.(type) {
	case nil:
	case Predicate:

		_, ok := exp.left.(Predicate)
		if ok {
			b.sb.WriteByte(' ')
			b.sb.WriteByte('(')
		}

		if err := b.buildExpression(exp.left, m); err != nil {
			return err
		}

		if ok {
			b.sb.WriteByte(' ')
			b.sb.WriteByte(')')
		}

		b.sb.WriteByte(' ')
		b.sb.WriteString(exp.op.String())

		_, ok = exp.right.(Predicate)
		if ok {
			b.sb.WriteByte(' ')
			b.sb.WriteByte('(')
		}

		if err := b.buildExpression(exp.right, m); err != nil {
			return err
		}

		if ok {
			b.sb.WriteByte(' ')
			b.sb.WriteByte(')')
		}

	case Column:
		f, ok := m.fields[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}

		b.sb.WriteByte(' ')
		b.sb.WriteByte('`')
		b.sb.WriteString(f.colName)
		b.sb.WriteByte('`')
	case value:
		b.sb.WriteByte(' ')
		b.sb.WriteByte('?')
		b.addArg(exp.val)
	}
	return nil
}

func (b *builder) addArg(args ...any) {
	if b.args == nil {
		b.args = make([]any, 0, len(args))
	}
	b.args = append(b.args, args...)
}
