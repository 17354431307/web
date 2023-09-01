package reflect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIterateArrayOrSlice(t *testing.T) {
	testCases := []struct {
		name string

		entity   any
		wantVals []any
		wantErr  error
	}{
		{
			name: "[3]int",

			entity:   [3]int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
		{
			name: "[]int",

			entity:   []int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := IterateArrayOrSlice(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVals, val)
		})
	}
}

func TestIterateMap(t *testing.T) {
	testCases := []struct {
		name string

		entity   any
		wantKeys []any
		wantVals []any
		wantErr  error
	}{
		{
			name: "map",
			entity: map[string]string{
				"A": "a",
				"B": "b",
			},
			wantKeys: []any{"A", "B"},
			wantVals: []any{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys, vals, err := IterateMap(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}

			assert.EqualValues(t, tc.wantKeys, keys)
			assert.EqualValues(t, tc.wantVals, vals)
		})
	}
}
