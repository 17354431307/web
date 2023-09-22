package orm

import "database/sql"

// Result 相当于 sql.Result 的一个装饰器
type Result struct {
	err error
	res sql.Result
}

func (r Result) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}

	return r.res.LastInsertId()
}

func (r Result) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}

	return r.res.RowsAffected()
}

func (r Result) Err() error {
	return r.err
}
