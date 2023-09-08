package orm

import (
	"context"
	"database/sql"
)

// Querier 用于 select 语句
type Querier[T any] interface {

	// 这种设计形态也是可以的
	//Get(ctx context.Context) (*T, error)
	//GetMulti(ctx context.Context) ([]T, error)

	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor 用于 insert, delete 和 update
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}

type QueryBuilder interface {
	Build() (*Query, error)
}

type Query struct {
	SQL  string
	Args []any
}

type TableName interface {
	TableName() string
}
