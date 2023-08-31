package orm_test

import (
	"database/sql"
	"github.com/Moty1999/web/orm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_Build(t *testing.T) {
	testCase := []struct {
		name string

		builder orm.QueryBuilder

		wantQuery *orm.Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: &orm.Selector[TestModel]{},
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
		{
			name:    "from",
			builder: (&orm.Selector[TestModel]{}).From("`test_model`"),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: (&orm.Selector[TestModel]{}).From(""),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
		{
			name:    "with db",
			builder: (&orm.Selector[TestModel]{}).From("`test_db`.`test_model`"),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `test_db`.`test_model`;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: (&orm.Selector[TestModel]{}).Where(orm.C("Age").Eq(18)),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `Age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: (&orm.Selector[TestModel]{}).Where(orm.Not(orm.C("Age").Eq(18))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel` WHERE NOT (`Age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: (&orm.Selector[TestModel]{}).Where(orm.C("Age").Eq(18).And(orm.C("FirstName").Eq("Tom"))),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`Age` = ?) AND (`FirstName` = ?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "empty where",
			builder: (&orm.Selector[TestModel]{}).Where(),
			wantQuery: &orm.Query{
				SQL:  "SELECT * FROM `TestModel`;",
				Args: nil,
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
