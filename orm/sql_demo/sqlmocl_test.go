package sql_demo

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestSqlMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	defer db.Close()
	assert.NoError(t, err)

	mackRows := sqlmock.NewRows([]string{"id", "first_name"})
	mackRows.AddRow(1, "Tom")

	// 正则表达式, 而且 mock 的位置也有关系
	mock.ExpectQuery("SELECT `id`, `first_name` FROM `user`.*").WillReturnRows(mackRows)
	mock.ExpectQuery("SELECT `id` FROM `user`.*").WillReturnError(errors.New("mock error"))

	rows, err := db.QueryContext(context.Background(), "SELECT `id`, `first_name` FROM `user` WHERE `id` = 1")
	assert.NoError(t, err)
	for rows.Next() {
		tm := &TestModel{}

		err = rows.Scan(&tm.Id, &tm.FirstName)
		assert.NoError(t, err)
		log.Println(tm)
	}

	_, err = db.QueryContext(context.Background(), "SELECT `id` FROM `user` WHERE `id` = 1")
	assert.NoError(t, err)

}
