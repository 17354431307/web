package orm

import (
	"github.com/Moty1999/web/orm/internal/errs"
	"github.com/Moty1999/web/orm/model"
	"reflect"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

type OnDuplicateKey struct {
	assigns []Assignable
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &OnDuplicateKey{
		assigns: assigns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	builder
	values  []*T
	db      *DB
	columns []string

	//onDuplicateKey []Assignable
	onDuplicateKey *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
	}
}

//func (i *Inserter[T]) OnDuplicateKey(vals ...Assignable) *Inserter[T] {
//	i.onDuplicateKey = vals
//	return i
//}

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

// 指定插入的数据
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}

	i.sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.builder.model = m

	// 拼接表名
	i.quote(m.TableName)

	i.sb.WriteByte('(')

	fields := m.Fields
	// 用户指定了列名
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, col := range i.columns {
			fdMeta, ok := m.FieldMap[col]
			if !ok {
				return nil, errs.NewErrUnknowField(col)
			}

			fields = append(fields, fdMeta)
		}
	}

	// 一定要显示指定列的顺序，不然我们不知道数据库中默认的顺序
	// 我们要构造 `test_model`(col1, col2...)
	// 不能遍历这个 FieldMap，ColMap，因为在 Go 里面 map 的遍历，每一次的顺序都不一样
	// 所以额外引入了一个 Fields 这个切片

	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(field.ColName)
	}
	i.sb.WriteByte(')')

	// 拼接 Values
	i.sb.WriteString(" VALUES ")

	// 预估的参数数量是：我有多少行乘以我有多少个字段
	i.args = make([]any, 0, len(i.values)*len(fields))
	for j, val := range i.values {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		for idx, field := range fields {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')

			// 把参数读出来
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			i.args = append(i.args, arg)
		}
		i.sb.WriteByte(')')
	}

	if i.onDuplicateKey != nil {
		err = i.dialect.buildOnDuplicateKey(&i.builder, i.onDuplicateKey)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')
	return &Query{SQL: i.sb.String(), Args: i.args}, nil
}
