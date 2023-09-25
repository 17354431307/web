package valuer

import (
	"database/sql"
	"github.com/Moty1999/web/orm/model"
)

type Value interface {
	Field(name string) (any, error)
	SetColumns(rows *sql.Rows) error
}

// Creator 函数式的工厂方法
type Creator func(model *model.Model, entity any) Value
