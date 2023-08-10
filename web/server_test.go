package web

import (
	"fmt"
	"net/http"
	"testing"
)

// server 就是 handler,  http 和 web 框架的结合点
func TestServer(t *testing.T) {
	h := &HTTPServer{}

	handler1 := func(ctx Context) {
		fmt.Println("处理第一件事")
	}
	handler2 := func(ctx Context) {
		fmt.Println("处理第二件事")
	}

	// 用户自己去管
	h.AddRoute(http.MethodGet, "/user", func(ctx Context) {
		handler1(ctx)
		handler2(ctx)
	})
	h.Get("/user", func(ctx Context) {

	})
	//h.AddRoute1(http.MethodGet, "/user", handler1, handler2)

	// 用法一, 完全委托给 http 包
	http.ListenAndServe(":8081", h)
	http.ListenAndServeTLS(":8081", "", "", h)

	// 用户二, 手动管
	h.Start(":8081")

}
