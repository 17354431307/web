package reflect

import (
	"errors"
	"reflect"
)

// InterateFields 遍历字段
// 这里只能接收 XXX 之类的数据
func InterateFields(entity any) (map[string]any, error) {

	if entity == nil {
		return nil, errors.New("不支持 nil")
	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if val.IsZero() {
		return nil, errors.New("不支持零值")
	}

	// 在这里用 for 的原因是, 如果是多级指针的话, 必须找到最后的指针指向的结构体
	for typ.Kind() == reflect.Pointer {
		// 拿到指针指向的对象
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("不支持的类型")
	}

	numField := typ.NumField()
	res := make(map[string]any, numField)
	for i := 0; i < numField; i++ {
		// 字段的类型
		fieldType := typ.Field(i)
		// 字段的值
		fieldValue := val.Field(i)
		if fieldType.IsExported() {
			res[fieldType.Name] = fieldValue.Interface()
		} else {
			res[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}

	}

	return res, nil
}

func SetField(entity any, field string, newValue any) error {
	val := reflect.ValueOf(entity)

	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	fieldName := val.FieldByName(field)
	if !fieldName.CanSet() {
		return errors.New("不可修改字段")
	}
	fieldName.Set(reflect.ValueOf(newValue))
	return nil
}
