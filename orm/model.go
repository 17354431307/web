package orm

import (
	"github.com/Moty1999/web/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	tagKeyColumn = "column"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOption) (*Model, error)
}

type ModelOption func(m *Model) error

type Model struct {
	tableName string
	fields    map[string]*Field
}

type Field struct {
	// 字段名
	GoName string
	// 列名
	colName string
	// 代表的是字段的类型
	typ reflect.Type
}

// register 代表着元数据的注册中心
type registry struct {
	// 读写锁
	//lock   sync.RWMutex
	models sync.Map
}

func newRegistry() *registry {
	return &registry{}
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}

	m, err := r.Register(val)
	if err != nil {
		return nil, err
	}
	return m.(*Model), nil
}

//func (r *registry) get1(val any) (*Model, error) {
//	/*
//		这里使用锁是使用 double check 的写法
//		检查了两遍
//	*/
//
//	typ := reflect.TypeOf(val)
//
//	r.lock.RLock()
//	m, ok := r.models[typ]
//	r.lock.RUnlock()
//
//	if ok {
//		return m, nil
//	}
//
//	r.lock.Lock()
//	defer r.lock.Unlock()
//	m, ok = r.models[typ]
//	if ok {
//		return m, nil
//	}
//
//	m, err := r.Register(val)
//	if err != nil {
//		return nil, err
//	}
//	r.models[typ] = m
//	return m, nil
//}

// Register 限制只能用一级指针
func (r *registry) Register(entity any, opts ...ModelOption) (*Model, error) {
	typ := reflect.TypeOf(entity)

	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}

	elemType := typ.Elem()
	numField := elemType.NumField()
	fieldMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fd := elemType.Field(i)
		pair, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		columnName := pair[tagKeyColumn]
		if columnName == "" {
			// 用户没有设置自定义列名
			columnName = underscoreName(fd.Name)
		}

		fieldMap[fd.Name] = &Field{
			colName: columnName,
			// 字段类型
			typ: fd.Type,
			// 字段名
			GoName: fd.Name,
		}
	}

	var tableName string
	if v, ok := entity.(TableName); ok {
		tableName = v.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemType.Name())
	}

	res := &Model{
		tableName: tableName,
		fields:    fieldMap,
	}

	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}

	r.models.Store(typ, res)
	return res, nil
}

func ModelWithTableName(tableName string) ModelOption {
	return func(m *Model) error {
		m.tableName = tableName
		return nil
	}
}

func ModelWithColumnName(field string, colName string) ModelOption {
	return func(m *Model) error {

		fd, ok := m.fields[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}

		fd.colName = colName
		return nil
	}
}

type User struct {
	ID int64 `orm:"column=id, xxx=bbb"`
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok {
		return map[string]string{}, nil
	}

	pairs := strings.Split(ormTag, ",")
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		pair := strings.Trim(pair, " ")
		segs := strings.Split(pair, "=")
		if len(segs) != 2 {
			return nil, errs.NewErrInvalidTagContext(pair)
		}
		key, val := segs[0], segs[1]
		res[key] = val
	}
	return res, nil
}

// underscoreName 驼峰转字符串命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}

	return string(buf)
}
