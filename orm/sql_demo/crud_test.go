package sql_demo

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestDB(t *testing.T) {
	dsn := "file:test.db?cache=shared&mode=memory"
	db, err := sql.Open("sqlite3", dsn)
	assert.NoError(t, err)
	defer db.Close()
	db.Ping()
	// 这里你就可以用 db 了

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

	// 除了 SELECT 语句，都是使用 ExecContext
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)

	// 完成了建表
	assert.NoError(t, err)

	// 使用 ? 作为查询的参数的占位符
	res, err := db.ExecContext(ctx,
		"INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`) VALUES (?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")

	assert.NoError(t, err)

	affected, err := res.RowsAffected()
	assert.NoError(t, err)
	log.Println("受影响的行数", affected)

	lastInsertId, err := res.LastInsertId()
	assert.NoError(t, err)
	log.Println("最后插入的Id", lastInsertId)

	row := db.QueryRowContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?;", 1)
	assert.NoError(t, err)
	tm := TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	assert.NoError(t, err)

	row = db.QueryRowContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?;", 2)
	tm = TestModel{}
	// 查询不到返回一个错误
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
	assert.ErrorIs(t, sql.ErrNoRows, err)

	rows, err := db.QueryContext(ctx, "SELECT `id`, `first_name`, `age`, `last_name` FROM `test_model` WHERE `id` = ?;", 1)

	// 查询不到返回一个错误
	for rows.Next() {
		tm = TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
		assert.NoError(t, err)
		log.Println(tm)
	}

	cancel()
}

func TestTX(t *testing.T) {
	dsn := "file:test.db?cache=shared&mode=memory"
	db, err := sql.Open("sqlite3", dsn)
	assert.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

	// 除了 SELECT 语句，都是使用 ExecContext
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)
	assert.NoError(t, err)

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	assert.NoError(t, err)

	// 使用 ? 作为查询的参数的占位符
	res, err := tx.ExecContext(ctx,
		"INSERT INTO test_model(`id`, `first_name`, `age`, `last_name`) VALUES (?, ?, ?, ?)",
		1, "Tom", 18, "Jerry")
	if err != nil {
		// 提交事务
		err = tx.Rollback()
		if err != nil {
			log.Println(err)
		}
		return
	}
	assert.NoError(t, err)

	affected, err := res.RowsAffected()
	assert.NoError(t, err)
	log.Println("受影响的行数", affected)

	// 提交错误
	err = tx.Commit()

	cancel()
}

func TestPrepareStatement(t *testing.T) {
	dsn := "file:test.db?cache=shared&mode=memory"
	db, err := sql.Open("sqlite3", dsn)
	assert.NoError(t, err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

	// 除了 SELECT 语句，都是使用 ExecContext
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS test_model(
    id INTEGER PRIMARY KEY,
    first_name TEXT NOT NULL,
    age INTEGER,
    last_name TEXT NOT NULL
)
`)

	stmt, err := db.PrepareContext(ctx, "SELECT * From `test_model` WHERE `id` = ?")
	assert.NoError(t, err)

	rows, err := stmt.QueryContext(ctx, 1)
	assert.NoError(t, err)

	for rows.Next() {
		tm := &TestModel{}

		err = rows.Scan(&tm.Id, &tm.FirstName, tm.Age, tm.LastName)
		assert.NoError(t, err)
		log.Println(tm)
	}
	cancel()
	// 整个应用关闭的时候调用
	stmt.Close()

	//stmt, err := db.PrepareContext(ctx, "SELECT * From `test_model` WHERE `id` in (?, ?, ?)")
	//stmt, err := db.PrepareContext(ctx, "SELECT * From `test_model` WHERE `id` in (?, ?, ?, ?)")
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
