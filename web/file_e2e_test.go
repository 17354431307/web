package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

func TestUpload(t *testing.T) {
	fmt.Println(os.Getwd())
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	assert.NoError(t, err)

	engine := &GoTemplateEngine{
		T: tpl,
	}
	h := NewHTTPServer(ServerWithTemplateEngine(engine))
	h.Get("/upload", func(ctx *Context) {
		err := ctx.Render("upload.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})

	fu := FileUploader{
		FileFiled: "myfile",

		DstPathFunc: func(header *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", header.Filename)
		},
	}
	h.Post("/upload", fu.Handle())
	h.Start(":8081")
}

func TestDownload(t *testing.T) {
	h := NewHTTPServer()

	fu := FileDownloader{
		Dir: filepath.Join("testdata", "download"),
	}
	h.Get("/download", fu.Handle())
	h.Start(":8081")
}

func TestStaticResource(t *testing.T) {
	h := NewHTTPServer()

	dir := filepath.Join("testdata", "static")
	s, err := NewStaticResourceHandler(dir)
	if err != nil {
		log.Println(err)
		return
	}

	// localhost:8081/static/xxx.jpg
	h.Get("/static/:file", s.Handle)
	h.Start(":8081")
}
