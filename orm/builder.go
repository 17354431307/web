package orm

import (
	"github.com/Moty1999/web/orm/internal/errs"
	"github.com/Moty1999/web/orm/model"
	"strings"
)

type builder struct {
	sb      strings.Builder
	args    []any
	model   *model.Model
	dialect Dialect
	quoter  byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) buildColumn(name string) error {
	fdMeta, ok := b.model.FieldMap[name]
	if !ok {
		return errs.NewErrUnknowField(name)
	}

	b.quote(fdMeta.ColName)
	return nil
}

func (b *builder) addArg(args ...any) {
	if b.args == nil {
		b.args = make([]any, 0, len(args))
	}
	b.args = append(b.args, args...)
}

func (b *builder) buildPredicates(ps []Predicate, m *model.Model) error {

	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p.And(ps[i])
	}

	return b.buildExpression(p, m)
}

func (b *builder) buildExpression(expr Expression, m *model.Model) error {

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
		f, ok := m.FieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknowField(exp.name)
		}

		b.sb.WriteByte(' ')
		b.sb.WriteByte('`')
		b.sb.WriteString(f.ColName)
		b.sb.WriteByte('`')
	case value:
		b.sb.WriteByte(' ')
		b.sb.WriteByte('?')
		b.addArg(exp.val)
	}
	return nil
}
