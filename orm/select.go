package orm

import (
	"context"
	"github.com/Moty1999/web/orm/internal/errs"
)

// Selectable 是一个标记接口
// 它代表的是查找的列, 或者聚合函数等
// SELECT
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	table string
	where []Predicate

	cols    []Selectable
	groupBy []Column
	having  []Predicate
	offset  int
	limit   int
	orderBy []OrderBy

	sess Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},

		sess: sess,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT ")

	err = s.buildColumns()
	if err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")

	if s.table == "" {
		// 我怎么拿到表名
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.TableName)
		s.sb.WriteByte('`')
	} else {
		//segs := strings.Split(s.table, ".")
		//sb.WriteByte('`')
		//sb.WriteString(segs[0])
		//sb.WriteByte('`')
		//sb.WriteByte('.')
		//sb.WriteByte('`')
		//sb.WriteString(segs[1])
		//sb.WriteByte('`')

		s.sb.WriteString(s.table)
	}

	// 构造 where 部分
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err = s.buildExpression(p); err != nil {
			return nil, err
		}
	}

	// 构造 group by 部分
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, c := range s.groupBy {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			if err = s.buildColumn(c); err != nil {
				return nil, err
			}
		}
	}

	// 构建 having 部分
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING")
		if err = s.buildPredicates(s.having, s.model); err != nil {
			return nil, err
		}
	}

	// 构建 order by 部分
	if len(s.orderBy) > 0 {
		s.sb.WriteString(" ORDER BY ")
		if err = s.buildOrderBy(); err != nil {
			return nil, err
		}
	}

	// 构建 limit 部分
	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArg(s.limit)
	}

	// 构建 offset 部分
	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArg(s.offset)
	}

	s.sb.WriteByte(';')

	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildExpression(expr Expression) error {

	switch e := expr.(type) {
	case nil:

	case Predicate:
		// 在这里处理p
		// p.left 构建好
		// p.op 构建好
		// p.right 构建好
		_, ok := e.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(e.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		if e.left != nil && e.op != "" {
			s.sb.WriteByte(' ')
		}

		if e.op != "" {
			s.sb.WriteString(e.op.String())
			s.sb.WriteByte(' ')
		}

		_, ok = e.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(e.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
	case Column:
		// 这种写法很隐晦
		e.alias = ""
		return s.buildColumn(e)
	case value:
		s.sb.WriteByte('?')
		s.addArg(e.val)
	case RawExpr:
		s.sb.WriteByte('(')
		s.sb.WriteString(e.raw)
		s.addArg(e.args...)
		s.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (s *Selector[T]) buildColumns() error {

	if len(s.cols) == 0 {
		// 没有指定列
		s.sb.WriteByte('*')
		return nil
	}

	for i, col := range s.cols {
		if i > 0 {
			s.sb.WriteString(", ")
		}

		switch c := col.(type) {
		case Column:
			err := s.buildColumn(c)
			if err != nil {
				return err
			}
		case Aggregate:
			// 聚合函数名
			s.sb.WriteString(c.fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(Column{name: c.arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
			// 聚合函数本身的别名
			if c.alias != "" {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(c.alias)
				s.sb.WriteByte('`')
			}
		case RawExpr:
			s.sb.WriteString(c.raw)
			s.addArg(c.args...)

		}

	}
	return nil
}

func (s *Selector[T]) buildColumn(c Column) error {
	// buildColumn(c Column, useAlias bool) 还有这样一种设计，使用 useAlias 这个标记位控制 Column 是否使用别名
	// 主要是为了处理在 Where 表达式式处理用户使用 Column 别名的错误用法

	fd, ok := s.model.FieldMap[c.name]
	// 字段不同, 或者说列不对
	if !ok {
		return errs.NewErrUnknowField(c.name)
	}

	s.sb.WriteByte('`')
	s.sb.WriteString(fd.ColName)
	s.sb.WriteByte('`')

	if c.alias != "" {
		s.sb.WriteString(" AS `")
		s.sb.WriteString(c.alias)
		s.sb.WriteByte('`')
	}
	return nil
}

func (s *Selector[T]) addArg(vals ...any) {
	if len(vals) == 0 {
		return
	}

	if s.args == nil {
		s.args = make([]any, 0, 8)
	}

	s.args = append(s.args, vals...)

}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

// func (s *Selector[T]) Query(query string, args ...any) *Selector[T] {
//
// }

// Where 这种形态的做法
func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

//func (s *Selector[T]) GetV1(ctx context.Context) (*T, error) {
//	query, err := s.Build()
//	if err != nil {
//		return nil, err
//	}
//
//	db := s.db.db
//	// 在这里，就是要发起查询，并且处理结果集
//	rows, err := db.QueryContext(ctx, query.SQL, query.Args...)
//	// 这个是查询错误
//	if err != nil {
//		return nil, err
//	}
//
//
//	// 你要确认有没有数据
//	if !rows.Next() {
//		// 要不要返回一个 error ？
//		// 返回 error，
//		return nil, ErrNoRows
//	}
//}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}

	// 在这里，就是要发起查询，并且处理结果集
	rows, err := s.sess.queryContext(ctx, query.SQL, query.Args...)
	// 这个是查询错误
	if err != nil {
		return nil, err
	}

	// 你要确认有没有数据
	if !rows.Next() {
		// 要不要返回一个 error ？
		// 返回 error，
		return nil, ErrNoRows
	}

	tp := new(T)
	val := s.creator(s.model, tp)

	err = val.SetColumns(rows)
	if err != nil {
		return nil, err
	}

	// 接口定义好之后, 就两件事情, 一个是用新接口的方法改造上层,
	// 一个就是提供不同的实现
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}

	// 在这里，就是要发起查询，并且处理结果集
	rows, err := s.sess.queryContext(ctx, query.SQL, query.Args)
	if err != nil {
		return nil, err
	}

	for !rows.Next() {
		// 要不要返回 error ?
		// 返回 error, 和 sql 包语义保持一致
		return nil, ErrNoRows
	}

	panic("implement me!")
}

//func (s *Selector[T]) Select(cols ...string) *Selector[T] {
//	s.cols = cols
//	return s
//}
//
//// 这种也是可行的
//// s.Select("first_name, last_name")
//func (s *Selector[T]) SelectV1(cols ...string) *Selector[T] {
//	s.cols = cols
//	return s
//}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.cols = cols
	return s
}

// GroupBy 设置 group by 子句
func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) OrderBy(orderBys ...OrderBy) *Selector[T] {
	s.orderBy = orderBys
	return s
}

func (s *Selector[T]) buildOrderBy() error {
	for idx, ob := range s.orderBy {
		if idx > 0 {
			s.sb.WriteByte(',')
		}
		err := s.buildColumn(Column{name: ob.col})
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(ob.order)
	}
	return nil
}

type OrderBy struct {
	col   string
	order string
}

func Asc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: "ASC",
	}
}

func Desc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: "DESC",
	}
}
