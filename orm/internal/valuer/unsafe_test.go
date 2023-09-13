package valuer

import "testing"

func TestUnsafeValue_SetColumns(t *testing.T) {
	testSetColumns(t, NewUnsafeValue)
}
