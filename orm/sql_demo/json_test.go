package sql_demo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonColumn_Value(t *testing.T) {

}

func TestJsonColumn_Scan(t *testing.T) {
	tsetCases := []struct {
		name    string
		src     any
		wantErr error
		wantVal User
		valid   bool
	}{
		{
			name: "nil",
		},
		{
			name: "string",
			src:  `{"Name": "Tom"}`,
			wantVal: User{
				Name: "Tom",
			},
			valid: true,
		},
		{
			name: "bytes",
			src:  []byte(`{"Name": "Tom"}`),
			wantVal: User{
				Name: "Tom",
			},
			valid: true,
		},
	}

	for _, tc := range tsetCases {
		t.Run(tc.name, func(t *testing.T) {
			js := &JsonColumn[User]{}
			err := js.Scan(tc.src)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, tc.wantVal, js.Val)
			assert.Equal(t, tc.valid, js.Valid)
		})
	}
}

type User struct {
	Name string
}
