package orm

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Moty1999/web/orm/internal/errs"
	"github.com/Moty1999/web/orm/internal/valuer"
	"github.com/Moty1999/web/orm/model"
)

type DBOption func(db *DB)

// DB 是一个 sql.DB 的装饰器
type DB struct {
	core
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
		core: core{
			r:       model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: DialectMySQL,
		},
		db: db,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(db *DB) {
		db.dialect = dialect
	}
}

func DBWithRegister(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
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

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx: tx,
	}, err
}

type txKey struct{}

func (db *DB) BeginTxV2(ctx context.Context, opts *sql.TxOptions) (context.Context, *Tx, error) {
	val := ctx.Value(txKey{})
	tx, ok := val.(*Tx)

	// 存在一个事务, 并且这个事务没有被提交或者回滚
	if ok && !tx.done {
		return ctx, tx, nil
	}

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	ctx = context.WithValue(ctx, txKey{}, tx)
	return ctx, tx, nil
}

// BeginTxV3 要求前面的人一定要开好事务
func (db *DB) BeginTxV3(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	val := ctx.Value(txKey{})
	tx, ok := val.(*Tx)
	if ok {
		return tx, nil
	}

	return nil, errors.New("没有开事务")
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) DoTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error, opts *sql.TxOptions) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	paniced := true
	defer func() {

		// panic 这里要设置好, 是个关键
		if paniced || err != nil {
			e := tx.Rollback()
			// 为什么这里要封装错误, 因为这里有3种错误, 业务逻辑错误, 回滚出现错误, 发生 panic
			// 为了不丢失错误信息, 所以需要包装错误
			err = errs.NewErrFailedToRollbackTx(err, e, paniced)
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(ctx, tx)
	paniced = false
	return err
}

func (db *DB) getCore() core {
	return db.core
}
