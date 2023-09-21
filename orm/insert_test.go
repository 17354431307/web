package orm

import (
	"database/sql"
	"github.com/Moty1999/web/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		i         QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			// 一行都没有
			name:    "no row",
			i:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			// 只插入一行
			name: "single row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?);",
				Args: []any{int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true}},
			},
		},
		{
			// 插入多行
			name: "multiple row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "He",
				Age:       19,
				LastName:  &sql.NullString{String: "Wen", Valid: true},
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{
					int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(13), "He", int8(19), &sql.NullString{String: "Wen", Valid: true},
				},
			},
		},
		{
			// 插入多行, 部分列
			name: "partial columns",
			i: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "He",
				Age:       19,
				LastName:  &sql.NullString{String: "Wen", Valid: true},
			}),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`) VALUES (?,?),(?,?);",
				Args: []any{
					int64(12), "Tom",
					int64(13), "He",
				},
			},
		},
		{
			name: "upsert-update value",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}).OnDuplicateKey().Update(Assign("FirstName", "Deng"), Assign("Age", 19)),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE " +
					"`first_name`=?,`age`=?;",
				Args: []any{
					int64(12),
					"Tom",
					int8(18),
					&sql.NullString{String: "Jerry", Valid: true},
					"Deng",
					19,
				},
			},
		},
		{
			name: "upset-update column",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        12,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{String: "Jerry", Valid: true},
			}, &TestModel{
				Id:        13,
				FirstName: "He",
				Age:       19,
				LastName:  &sql.NullString{String: "Wen", Valid: true},
			}).OnDuplicateKey().Update(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`,`first_name`,`age`,`last_name`) VALUES (?,?,?,?),(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`age`=VALUES(`age`);",
				Args: []any{
					int64(12), "Tom", int8(18), &sql.NullString{String: "Jerry", Valid: true},
					int64(13), "He", int8(19), &sql.NullString{String: "Wen", Valid: true},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
