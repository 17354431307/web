package file_demo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFile(t *testing.T) {
	//f, err := os.Open("testdata/my_file.txt")
	//assert.NoError(t, err)
	//
	//data := make([]byte, 1<<10)
	//n, err := f.Read(data)
	//fmt.Println(n)
	//assert.NoError(t, err)

	fmt.Println(os.Getwd())
	f, err := os.OpenFile("testdata/my_file.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	assert.NoError(t, err)
	n, err := f.WriteString("小文真帅!")
	fmt.Println(n)
	assert.NoError(t, err)
	f.Close()

	f, err = os.Create("testdata/my_file_copy.txt")
	assert.NoError(t, err)
	_, err = f.WriteString("Hello, World!")
	assert.NoError(t, err)
}
