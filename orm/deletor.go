package orm

import "strings"

type Deletor[T any] struct {
	tableName string
	where     []Predicate
	model     *model
	builder
	db *DB
}

func NewDeletor[T any](db *DB) *Deletor[T] {
	return &Deletor[T]{
		db: db,
	}
}

func (d *Deletor[T]) From(tableName string) *Deletor[T] {
	d.tableName = tableName
	return d
}

func (d *Deletor[T]) Where(ps ...Predicate) *Deletor[T] {
	d.where = append(d.where, ps...)
	return d
}

func (d *Deletor[T]) Build() (*Query, error) {
	if d.sb == nil {
		d.sb = &strings.Builder{}
	}

	var err error
	d.model, err = d.db.r.get(new(T))
	if err != nil {
		return nil, err
	}

	d.sb.WriteString("DELETE FROM")

	if d.tableName != "" {
		d.sb.WriteByte(' ')
		d.sb.WriteString(d.tableName)
	} else {
		d.sb.WriteByte(' ')
		d.sb.WriteByte('`')
		d.sb.WriteString(d.model.tableName)
		d.sb.WriteByte('`')
	}

	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE")
		if err := d.buildPredicates(d.where, d.model); err != nil {
			return nil, err
		}
	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, err
}

// 设计形式一
/*type Predicates []Predicate

func (ps Predicates) build(s *strings.Builder) error {
	// 写在这里
	panic("implement me!")
}*/

// 设计形式二
/*type predicates struct {
	// WHERE 或者 HAVING
	prefix string
	ps     []Predicate
}

func (ps predicates) build(s *strings.Builder) error {
	// 包含拼接 WHERE 或者 HAVING 的部分
	// 写在这里
}*/
