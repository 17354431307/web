package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"regexp"
	"testing"
)

func TestRouter_addRoute(t *testing.T) {
	// 第一步: 构造路由树
	// 第二步: 验证路由树
	testRouter := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail/:id",
		},
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	var mockHandler HandleFunc = func(ctx *Context) {

	}
	r := NewRouter()
	for _, route := range testRouter {
		r.addRoute(route.method, route.path, mockHandler)
	}

	// 在这里断言路由树和你预期的一模一样
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"home": {
								path:    "home",
								handler: mockHandler,
							},
						},
						handler: mockHandler,
					},
					"order": {
						path: "order",
						children: map[string]*node{
							"detail": {
								path:    "detail",
								handler: mockHandler,
								paramsChild: &node{
									path:    ":id",
									handler: mockHandler,
								},
							},
						},
						starChild: &node{
							path:    "*",
							handler: mockHandler,
						},
					},
				},
				handler: mockHandler,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {
						path: "order",
						children: map[string]*node{
							"create": {
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": {
						path:    "login",
						handler: mockHandler,
					},
				},
			},
		},
	}

	// 断言两者相等

	// 这个是不行的, 因为 HandleFunc 是不可比的
	//assert.Equal(t, wantRouter, r)

	msg, ok := wantRouter.equal(&r)
	assert.True(t, ok, msg)

	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	}, "path 不能为空")

	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	}, "web: 路径不能以 / 结尾")

	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a///c/", mockHandler)
	}, "web: 不能有连续的 /")

	r = NewRouter()
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	}, "web: 路由冲突, 重复注册[/]")

	r = NewRouter()
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	}, "web: 路由冲突, 重复注册[/a/b/c]")

	// 可用的 http method, 要不要检验
	// mockHandler 为 nil

	r = NewRouter()
	r.addRoute(http.MethodGet, "/a/*", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	}, "web: 不允许同时注册路径参数和通配符匹配, 已有通配符匹配")

	r = NewRouter()
	r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
	}, "web: 不允许同时注册路径参数和通配符匹配, 有参数路径匹配")

}

func TestRouter_findRoute(t *testing.T) {
	testRoute := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},

		{
			method: http.MethodGet,
			path:   "/order/*",
		},

		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		{
			method: http.MethodPost,
			path:   "/login/:username",
		},
	}

	var mockHandleFunc HandleFunc = func(ctx *Context) {}

	r := NewRouter()
	for _, route := range testRoute {
		r.addRoute(route.method, route.path, mockHandleFunc)
	}

	testCase := []struct {
		name   string
		method string
		path   string

		wantFound bool
		info      *matchInfo
	}{
		{
			// 方法不存在
			name:      "method not found",
			method:    http.MethodOptions,
			path:      "/order/detail",
			wantFound: false,
			info: &matchInfo{
				node: &node{
					handler: mockHandleFunc,
					path:    "detail",
				},
			},
		},
		{
			// 完全命中
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					handler: mockHandleFunc,
					path:    "detail",
				},
			},
		},
		{
			// 通配命中
			name:      "order start",
			method:    http.MethodGet,
			path:      "/order/abc",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					handler: mockHandleFunc,
					path:    "*",
				},
			},
		},
		{
			// 命中了, 但是没有 handler
			name:      "order",
			method:    http.MethodGet,
			path:      "/order",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path: "order",
					children: map[string]*node{
						"detail": &node{
							handler: mockHandleFunc,
							path:    "detail",
						},
					},
				},
			},
		},
		{
			// 根节点
			name:      "root",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path:    "/",
					handler: mockHandleFunc,
				},
			},
		},
		{
			// 没有注册路径
			name:   "path not found",
			method: http.MethodGet,
			path:   "/aaaabbbccc",
		},
		{
			// username 参数匹配
			name:      "login username",
			method:    http.MethodPost,
			path:      "/login/hexiaowen",
			wantFound: true,
			info: &matchInfo{
				node: &node{
					path:    ":username",
					handler: mockHandleFunc,
				},
				pathParams: map[string]string{
					"username": "hexiaowen",
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			info, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}

			assert.Equal(t, tc.info.pathParams, info.pathParams)
			msg, ok := tc.info.node.equal(info.node)
			assert.True(t, ok, msg)
		})
	}
}

func TestReg(t *testing.T) {
	reg := regexp.MustCompile(`:(\w+)(?:\((\w+)\))?`)
	data1 := ":id"
	ret1 := reg.FindStringSubmatch(data1)
	t.Log(ret1, len(ret1))

	data2 := ":(paramName"
	ret2 := reg.FindStringSubmatch(data2)
	t.Log(ret2)
}

// string 返回一个错误信息, 帮助我们排查问题
// bool 代表是否真的相等
func (r *router) equal(y *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的 http method"), false
		}
		// v, dst 要相等
		msg, ok := v.equal(dst)
		if !ok {
			return msg, false
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if n.path != y.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不相等"), false
	}

	if n.starChild != nil {
		msg, ok := n.starChild.equal(y.starChild)
		if !ok {
			return msg, ok
		}
	}

	if n.paramsChild != nil {
		msg, ok := n.paramsChild.equal(y.paramsChild)
		if !ok {
			return msg, ok
		}
	}

	// 比较 handler
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)
	if nHandler != yHandler {
		return fmt.Sprintf("handler 不相等"), false
	}

	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在", path), false
		}
		msg, ok := c.equal(dst)
		if !ok {
			return msg, ok
		}
	}

	return "", true
}
