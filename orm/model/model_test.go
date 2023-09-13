package model

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/Moty1999/web/orm/internal/errs"
	"github.com/stretchr/testify/assert"
)

func Test_registry_Register(t *testing.T) {
	testCases := []struct {
		name string

		entity    any
		wantModel *Model
		fields    []*Field
		wantErr   error
	}{
		{
			name:    "struct",
			entity:  TestModel{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Type:    reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Type:    reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Type:    reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
		},
		{
			name:    "map",
			entity:  map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "basic types",
			entity:  0,
			wantErr: errs.ErrPointerOnly,
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, f := range tc.fields {
				fieldMap[f.GoName] = f
				columnMap[f.ColName] = f
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name string

		entity    any
		wantModel *Model
		wantErr   error
		fields    []*Field
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Type:    reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Type:    reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "age",
					GoName:  "Age",
					Type:    reflect.TypeOf(int8(0)),
					Offset:  24,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Type:    reflect.TypeOf(&sql.NullString{}),
					Offset:  32,
				},
			},
		},
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column=first_name_t"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name_t",
					Type:    reflect.TypeOf(""),
				},
			},
		},
		{
			name: "empty column",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column="`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name",
					Type:    reflect.TypeOf(""),
				},
			},
		},
		{
			name: "column only",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewErrInvalidTagContext("column"),
		},
		{
			name: "ignore tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"abc=abc"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
			},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name",
					Type:    reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
			},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name",
					Type:    reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
			},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name",
					Type:    reflect.TypeOf(""),
				},
			},
		},
		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
			},
			fields: []*Field{
				{
					GoName:  "FirstName",
					ColName: "first_name",
					Type:    reflect.TypeOf(""),
				},
			},
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			fieldMap := make(map[string]*Field)
			columnMap := make(map[string]*Field)
			for _, f := range tc.fields {
				fieldMap[f.GoName] = f
				columnMap[f.ColName] = f
			}
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColumnMap = columnMap
			assert.Equal(t, tc.wantModel, m)

			typ := reflect.TypeOf(tc.entity)
			v, ok := r.(*registry).models.Load(typ)
			assert.True(t, ok)
			assert.Equal(t, tc.wantModel, v)
		})
	}
}

type CustomTableName struct {
	FirstName string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	FirstName string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

type EmptyTableName struct {
	FirstName string
}

func (c *EmptyTableName) TableName() string {
	return ""
}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name string

		entity        any
		tableName     string
		wantTableName string
	}{
		{
			name:          "empty table name",
			entity:        &TestModel{},
			tableName:     "",
			wantTableName: "",
		},
		{
			name:          "table name",
			entity:        &TestModel{},
			tableName:     "test_model_ttt",
			wantTableName: "test_model_ttt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			model, err := r.Register(tc.entity, ModelWithTableName(tc.tableName))
			assert.NoError(t, err)

			assert.Equal(t, tc.wantTableName, model.TableName)
		})
	}

}

func TestModelWithTableName(t *testing.T) {
	testCases := []struct {
		name        string
		field       string
		colName     string
		wantColName string
		wantErr     error
	}{
		{
			name:        "column name",
			field:       "FirstName",
			colName:     "first_name_ccc",
			wantColName: "first_name_ccc",
		},
		{
			name:    "invalid column name",
			field:   "XXX",
			colName: "first_name_ccc",
			wantErr: errs.NewErrUnknowField("XXX"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			model, err := r.Register(&TestModel{}, ModelWithColumnName(tc.field, tc.colName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			fd, ok := model.FieldMap[tc.field]
			assert.True(t, ok)
			assert.Equal(t, tc.wantColName, fd.ColName)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
