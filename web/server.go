package web

import (
	"fmt"
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

// TODO 采用的是 options 模式, 也是无侵入式的, 牛逼, 必须得吃透
type HTTPServerOption func(server *HTTPServer)

type HTTPServer struct {
	router
	mdls []Middleware

	log func(msg string, args ...any)
}

// 另外一种方案, 我不喜欢, 缺乏扩展性
//func NewHTTPServer(mdls ...Middleware) *HTTPServer {
//	res := &HTTPServer{
//		router: NewRouter(),
//		mdls:   mdls,
//	}
//	return res
//}

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: NewRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.mdls = mdls
	}
}

// ServeHTTP 处理请求的入口
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// 你的框架代码在这里
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	//h.serve(ctx)

	// 这里构造链条非常经典

	// 最后一个是这个
	root := h.serve

	// 然后这里就是利用最后一个不同往前回溯组装链条
	// 从后往前, 把后一个作为前一个的参数(next) 构造好链条
	for i := len(h.mdls) - 1; i >= 0; i-- {
		root = h.mdls[i](root)
	}

	// 这里执行的时候, 就是从前往后了
	// 这里最后一个步骤, 就是把 RespData 和 RespStatusCode 刷新到响应里面

	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			// 这一步就设置好了 RespData 和 RespStatusCode
			next(ctx)

			// 这里就相当于最后执行
			h.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (h *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		h.log("写入响应数据失败 %v", err)
	}
}

func (h *HTTPServer) serve(ctx *Context) {
	// 接下来查找路由并且执行命中的业务逻辑
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.node.handler == nil {
		// 路由没有命中
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("Not Found")

		return
	}

	ctx.PathParams = info.pathParams
	ctx.MatchedRoute = info.node.route
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
