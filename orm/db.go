package orm

import "database/sql"

type DBOption func(db *DB)

// DB 是一个 sql.DB 的装饰器
type DB struct {
	r  *registry
	db *sql.DB
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
		r:  newRegistry(),
		db: db,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	db, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}

	return db
}
