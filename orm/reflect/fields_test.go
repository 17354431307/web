package reflect

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestInterateFields(t *testing.T) {

	type User struct {
		Name string
		age  int
	}

	testCases := []struct {
		name string

		entity any

		wantRes map[string]any
		wantErr error
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				// age 是私有的, 拿不到, 最终我们创建了零值来填充
				"age": 0,
			},
		},
		{
			name: "pointer",
			entity: &User{
				Name: "Tom",
				age:  18,
			},
			wantRes: map[string]any{
				"Name": "Tom",
				// age 是私有的, 拿不到, 最终我们创建了零值来填充
				"age": 0,
			},
		},
		{
			name:    "basic type",
			entity:  18,
			wantErr: errors.New("不支持的类型"),
		},
		{
			name: "multiple pointer",
			entity: func() **User {
				res := &User{
					Name: "Tom",
					age:  18,
				}
				return &res
			}(),
			wantRes: map[string]any{
				"Name": "Tom",
				"age":  0,
			},
		},
		{
			name:    "nil",
			entity:  nil,
			wantErr: errors.New("不支持 nil"),
		},
		{
			name:    "user nil",
			entity:  (*User)(nil),
			wantErr: errors.New("不支持零值"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := InterateFields(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestSetField(t *testing.T) {
	type User struct {
		Name string
		age  int
	}

	testCases := []struct {
		name string

		entity any
		field  string
		newVal any

		wantErr error
		// 修改后的 entity
		wantEntity any
	}{
		{
			name: "struct",

			entity:  User{Name: "Tom"},
			field:   "Name",
			newVal:  "Jerry",
			wantErr: errors.New("不可修改字段"),
		},
		{
			name: "pointer",

			entity:     &User{Name: "Tom"},
			field:      "Name",
			newVal:     "Jerry",
			wantEntity: &User{Name: "Jerry"},
		},
		{
			name: "pointer exported",

			entity:  &User{age: 18},
			field:   "age",
			newVal:  20,
			wantErr: errors.New("不可修改字段"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetField(tc.entity, tc.field, tc.newVal)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantEntity, tc.entity)
		})
	}

	var i = 0
	ptr := &i
	reflect.ValueOf(ptr).Elem().SetInt(18)
	fmt.Printf("%d", i)

}
