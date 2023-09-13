package additional

import "errors"

type MyStructOption func(myStruct *MyStruct)
type MyStructOptionErr func(myStruct *MyStruct) error

type MyStruct struct {
	// 第一个部分是必须用户输入的字段

	id   uint64
	name string

	// 第二个部分是可选的字段
	address string

	// 这里可以有很多字段

	// field1 和 field2 要一起存在的
	field1 int
	filed2 int
}

func WithField1AndField2(field1, field2 int) MyStructOption {
	return func(myStruct *MyStruct) {
		myStruct.field1 = field1
		myStruct.filed2 = field2
	}
}

func WithAddressV2(address string) MyStructOption {
	return func(myStruct *MyStruct) {
		if address == "" {
			panic("地址不能为空字符串")
		}
		myStruct.address = address
	}
}

func WithAddressV1(address string) MyStructOptionErr {
	return func(myStruct *MyStruct) error {
		if address == "" {
			return errors.New("地址不能为空字符串")
		}
		myStruct.address = address
		return nil
	}
}

func WithAddress(address string) MyStructOption {
	return func(myStruct *MyStruct) {
		myStruct.address = address
	}
}

//var m = MyStruct{}

// NewNewMyStruct 参数包含用户必须输入的字段
func NewMyStruct(id uint64, name string, opts ...MyStructOption) *MyStruct {
	res := &MyStruct{
		id:   id,
		name: name,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func NewMyStructV1(id uint64, name string, opts ...MyStructOptionErr) (*MyStruct, error) {
	res := &MyStruct{
		id:   id,
		name: name,
	}

	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
