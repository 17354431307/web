package orm

import (
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
				tableName: "test_model",
				fields: map[string]*Field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
					},
					"Age": {
						colName: "age",
					},
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

	r := newRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

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
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				tableName: "test_model",
				fields: map[string]*Field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
					},
					"Age": {
						colName: "age",
					},
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
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name_t",
					},
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
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
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
				tableName: "tag_table",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name:   "table name",
			entity: &CustomTableName{},
			wantModel: &Model{
				tableName: "custom_table_name_t",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				tableName: "custom_table_name_ptr_t",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				tableName: "empty_table_name",
				fields: map[string]*Field{
					"FirstName": {
						colName: "first_name",
					},
				},
			},
		},
	}

	r := newRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)

			typ := reflect.TypeOf(tc.entity)
			v, ok := r.models.Load(typ)
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
			r := newRegistry()
			model, err := r.Register(tc.entity, ModelWithTableName(tc.tableName))
			assert.NoError(t, err)

			assert.Equal(t, tc.wantTableName, model.tableName)
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
			wantErr: errs.NewErrUnknownField("XXX"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := newRegistry()
			model, err := r.Register(&TestModel{}, ModelWithColumnName(tc.field, tc.colName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			fd, ok := model.fields[tc.field]
			assert.True(t, ok)
			assert.Equal(t, tc.wantColName, fd.colName)
		})
	}
}
