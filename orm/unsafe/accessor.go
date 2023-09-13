package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

/*
	前面我们使用了 unsafe.Pointer 和 uintptr, 这两者都代表者指针，那么有什么区别？

	+ unsafe.Pointer: 是 Go 层面的指针，Gc 会维护 unsafe.Pointer 的值
		比如说：一个对象 Gc 前放在 A 地方，Gc 后放在了 B 地方，Go 会帮你调整 unsafe.Pointer 的值 ，确保里面的值执行那个对象

	+ uintptr: 直接就是一个数字，代表的是一个内存地址
		但是 uintptr 是一个绝对的量，也就是说在 GC 前后指向的内存地址是不变的，那这就会导致在 GC 前后内存地址里值是会发生变化的
		但是这个还是有点用处的，比如表示相对量，偏移量，或者做地址的加减法的时候，uintptr 还是挺还用的

	总结：只在进行地址运算的时候使用 uintptr，其他的时候都用 unsafe.Pointer
*/

type UnsafeAccessor struct {
	fields  map[string]FieldMeta
	address unsafe.Pointer
}

func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	typ := reflect.TypeOf(entity)

	typ = typ.Elem()
	numField := typ.NumField()
	fields := make(map[string]FieldMeta, numField)

	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = FieldMeta{
			Offset: fd.Offset,
			typ:    fd.Type,
		}
	}

	// 通过反射来获取对象的起始地址
	val := reflect.ValueOf(entity)

	return &UnsafeAccessor{
		fields:  fields,
		address: val.UnsafePointer(),
	}
}

func (a *UnsafeAccessor) Field(field string) (any, error) {
	// 起始地址 + 偏移量
	meta, ok := a.fields[field]
	if !ok {
		return nil, errors.New("非法字段")
	}

	// 字段起始地址
	fdAddr := unsafe.Pointer(uintptr(a.address) + meta.Offset)

	// 如果知道类型，就这么读
	//return *(*int)(fdAddr), nil

	// 如果不知道类型，就这么读
	return reflect.NewAt(meta.typ, fdAddr).Elem().Interface(), nil
}

func (a *UnsafeAccessor) SetField(field string, val int) error {
	// 起始地址 + 偏移量

	meta, ok := a.fields[field]
	if !ok {
		return errors.New("非法字段")
	}

	// 字段真实地址
	fdAddr := unsafe.Pointer(uintptr(a.address) + meta.Offset)

	// 你知道确切类型
	//*(*int)(fdAddr) = val

	// 你不知道确切的类型
	reflect.NewAt(meta.typ, fdAddr).Elem().Set(reflect.ValueOf(val))
	return nil
}

type FieldMeta struct {
	Offset uintptr
	typ    reflect.Type
}
