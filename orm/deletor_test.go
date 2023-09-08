package orm

import (
	"github.com/Moty1999/web/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeletor_Builder(t *testing.T) {

	type TestModel struct {
		Id        string
		Name      string
		Age       int
		FirstName string
	}

	db := memoryDB(t)

	testCases := []struct {
		name string

		builder   QueryBuilder
		entity    any
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewDeletor[TestModel](db),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "from",
			builder: NewDeletor[TestModel](db).From("`test_db`.`test_model`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_db`.`test_model`;",
			},
		},
		{
			name:    "empty from",
			builder: NewDeletor[TestModel](db).From(""),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name:    "where",
			builder: NewDeletor[TestModel](db).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: NewDeletor[TestModel](db).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE NOT ( `age` = ? );",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: NewDeletor[TestModel](db).Where(C("Age").Eq(18).And(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model` WHERE ( `age` = ? ) AND ( `first_name` = ? );",
				Args: []any{
					18, "Tom",
				},
			},
		},
		{
			name:    "invalid column",
			builder: NewDeletor[TestModel](db).Where(C("XXXX").Eq(20)),
			wantErr: errs.NewErrUnknownField("XXXX"),
		},
		{
			name:    "empty where",
			builder: NewDeletor[TestModel](db).Where(),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, tc.wantQuery, query)
		})
	}
}
