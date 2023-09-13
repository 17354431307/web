package orm

import (
	"database/sql"
	"github.com/Moty1999/web/orm/internal/valuer"
	"github.com/Moty1999/web/orm/model"
)

type DBOption func(db *DB)

// DB 是一个 sql.DB 的装饰器
type DB struct {
	r       model.Registry
	db      *sql.DB
	creator valuer.Creator
}

func Open(driver string, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}

	return OpenDB(db, opts...)
}

// OpenDB 常用于测试，以及集成别的数据库中间件。我们会使用 sqlmock 来做单元测试
func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:       model.NewRegistry(),
		db:      db,
		creator: valuer.NewUnsafeValue,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	db, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}

	return db
}
