package web

import (
	"net"
	"net/http"
)

// 确保一定实现了 Server 接口
var _ Server = &HTTPServer{}

type HandleFunc func(ctx *Context)

// 核心 Server 抽象
type Server interface {
	http.Handler

	Start(addr string) error

	// addRoute 增加路由注册的功能
	// method 是 HTTP 方法
	// path 是路径
	// handleFunc 是业务逻辑
	addRoute(method string, path string, handleFunc HandleFunc)

	// 这种允许注册多个, 没有必要提供
	// 让用户自己去管
	//addRoute1(method string, path string, handlesFunc ...HandleFunc)
}

//type HTTPSServer struct {
//	HTTPServer
//}

type HTTPServer struct {
	router
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: NewRouter(),
	}
}

// ServerHTTP 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 你的框架代码在这里
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	h.serve(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	// 接下来查找路由并且执行命中的业务逻辑
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.node.handler == nil {
		// 路由没有命中
		ctx.Resp.WriteHeader(http.StatusNotFound)
		ctx.Resp.Write([]byte("NOT FOUND"))
		return
	}

	ctx.PathParams = info.pathParams
	info.node.handler(ctx)
}

//func (h *HTTPServer) addRoute(method string, path string, handleFunc HandleFunc) {
//	// 这里注册到路由树里面
//	//panic("implement me")
//}

// 为什么取别名, 防止用户乱传 method
func (h *HTTPServer) Get(path string, handleFunc HandleFunc) {
	// 委托给 addRoute 去执行, 这种用法很常见
	h.addRoute(http.MethodGet, path, handleFunc)
}

func (h *HTTPServer) Post(path string, handleFunc HandleFunc) {
	// 委托给 addRoute 去执行, 这种用法很常见
	h.addRoute(http.MethodPost, path, handleFunc)
}

func (h *HTTPServer) Options(path string, handleFunc HandleFunc) {
	// 委托给 addRoute 去执行, 这种用法很常见
	h.addRoute(http.MethodOptions, path, handleFunc)
}

//func (h *HTTPServer) addRoute1(method string, path string, handlesFunc ...HandleFunc) {
//	panic("implement me")
//}

func (h *HTTPServer) Start(addr string) error {
	// 也可以这种方式创建 Server
	//http.Server{}

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 在这里, 可以让用户注册所谓的 after start 回调
	// 比如说往你的 admin 注册一下自己这个实例
	// 在这里执行一些你业务所需的前置条件

	return http.Serve(listen, h)
}

//func (h *HTTPServer) Start1(addr string) error {
//	return http.ListenAndServe(addr, h)
//}
