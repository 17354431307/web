package template_demo

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	type User struct {
		Name string
	}
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{.Name}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, User{
		Name: "Tom",
	})
	require.NoError(t, err)
	assert.Equal(t, `Hello, Tom`, buffer.String())

}

func TestMapData(t *testing.T) {

	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{.Name}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, map[string]string{
		"Name": "Tom",
	})
	require.NoError(t, err)
	assert.Equal(t, `Hello, Tom`, buffer.String())

}

func TestSliceData(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{index . 0}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, []string{"Tom"})
	require.NoError(t, err)
	assert.Equal(t, `Hello, Tom`, buffer.String())
}

func TestBasic(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`Hello, {{.}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, 123)
	require.NoError(t, err)
	assert.Equal(t, `Hello, 123`, buffer.String())
}

func TestFuncCall(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
切片长度: {{len .Slice}}
{{printf "%.2f" 1.2345}}
Hello, {{.Hello "Tom" "Jerry"}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, FuncCall{
		Slice: []string{
			"a", "b",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, `
切片长度: 2
1.23
Hello, Tom-Jerry`, buffer.String())
}

func TestForLoop(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
{{- range $idx, $ele := .Slice}}
{{- .}}
{{$idx}}-{{$ele}}
{{end}}
`)
	assert.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, FuncCall{
		Slice: []string{"a", "b"},
	})
	assert.NoError(t, err)
	assert.Equal(t, `a
0-a
b
1-b

`, buffer.String())

}

func TestForLoopByNumber(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
{{- range $idx, $ele := . -}}
{{$idx}}
{{end}}
`)
	assert.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, make([]int, 3))
	assert.NoError(t, err)
	assert.Equal(t, `0
1
2

`, buffer.String())

}

func TestIfElse(t *testing.T) {
	type User struct {
		Age int
	}

	tpl := template.New("hello")
	tpl, err := tpl.Parse(`
{{- if and (gt .Age 0) (le .Age 6)}}
我是儿童: (0, 6]
{{else if and (gt .Age 6) (le .Age 18)}}
我是少年: (6, 18]
{{ else }}
我是成年: [18, +inf)
{{end -}}
`)
	assert.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, User{Age: 19})
	assert.NoError(t, err)
	assert.Equal(t, `
我是成年: [18, +inf)
`, buffer.String())
}

type FuncCall struct {
	Slice []string
}

func (f FuncCall) Hello(first string, last string) string {
	return fmt.Sprintf("%s-%s", first, last)
}
