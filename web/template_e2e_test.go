//go:build e2e

package web

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"log"
	"testing"
)

func TestLoginPage(t *testing.T) {
	// 不要这样子做
	// tplName = tplName + ".gohtml"
	// tplName = tplName + c.tplPrefix

	tpl, err := template.ParseGlob("testdata/tpls/*gohtml")
	assert.NoError(t, err)
	engine := &GoTemplateEngine{
		T: tpl,
	}
	h := NewHTTPServer(ServerWithTemplateEngine(engine))

	h.Get("/login", func(ctx *Context) {
		err := ctx.Render("login.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})
	h.Start(":8081")
}
