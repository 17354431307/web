package orm

import (
	"github.com/Moty1999/web/orm/internal/errs"
	"reflect"
	"strings"
)

type Inserter[T any] struct {
	values []*T
	db     *DB
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
	}
}

// 指定插入的数据
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}

	var sb strings.Builder

	sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}

	// 拼接表名
	sb.WriteByte('`')
	sb.WriteString(m.TableName)
	sb.WriteByte('`')

	// 一定要显示指定列的顺序，不然我们不知道数据库中默认的顺序
	// 我们要构造 `test_model`(col1, col2...)
	sb.WriteByte('(')
	// 不能遍历这个 FieldMap，ColMap，因为在 Go 里面 map 的遍历，每一次的顺序都不一样
	// 所以额外引入了一个 Fields 这个切片
	for idx, field := range m.Fields {
		if idx > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('`')
		sb.WriteString(field.ColName)
		sb.WriteByte('`')
	}
	sb.WriteByte(')')

	// 拼接 Values
	sb.WriteString(" VALUES ")

	// 预估的参数数量是：我有多少行乘以我有多少个字段
	args := make([]any, 0, len(i.values)*len(m.Fields))
	for j, val := range i.values {
		if j > 0 {
			sb.WriteByte(',')
		}
		sb.WriteByte('(')
		for idx, field := range m.Fields {
			if idx > 0 {
				sb.WriteByte(',')
			}
			sb.WriteByte('?')

			// 把参数读出来
			arg := reflect.ValueOf(val).Elem().FieldByName(field.GoName).Interface()
			args = append(args, arg)
		}
		sb.WriteByte(')')
	}

	sb.WriteByte(';')
	return &Query{SQL: sb.String(), Args: args}, nil
}
