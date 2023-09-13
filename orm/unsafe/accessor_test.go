package unsafe

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnsafeAccessor_Field(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	user := &User{
		Name: "Tom",
		Age:  20,
	}

	accessor := NewUnsafeAccessor(user)
	val, err := accessor.Field("Age")
	assert.NoError(t, err)
	assert.Equal(t, 20, val)

	err = accessor.SetField("Age", 19)
	assert.NoError(t, err)
	assert.Equal(t, 19, user.Age)
}
