package web

import (
	"fmt"
	"net/http"
	"testing"
)

// server 就是 handler,  http 和 web 框架的结合点
//func TestServer(t *testing.T) {
//
//}

// middleware 符合洋葱模型, 洋葱模型是可以往回走的责任链模式
// 洋葱模式用来无侵入式地增强核心功能, 或者解决 AOP 问题
func TestHTTPServer_ServeHTTP(t *testing.T) {
	server := NewHTTPServer()
	server.mdls = []Middleware{
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第一个before")
				next(ctx)
				fmt.Println("第一个after")
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第二个before")
				next(ctx)
				fmt.Println("第二个after")
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第三个中断")
			}
		},
		func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("第四个, 你看不到这句话")
			}
		},
	}
	server.ServeHTTP(nil, &http.Request{})
}
