package web

import (
	"fmt"
	"strings"
)

// 用来支持对路由树的操作
// 代表路由(森林)
type router struct {
	// Beego Gin HTTP method 对应一颗树
	// GET 有一棵树, POST 也有一颗树

	// http method => 路由树节点
	trees map[string]*node
}

func NewRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// 加一些限制:
// path 必须以 / 开头, 不能以 / 结尾, 中间也不能有连续的 //, 不能为空
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("path 不能为空")
	}

	root, ok := r.trees[method]
	if !ok {
		// 说明没有根节点
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	// 开头的检验
	if path[0] != '/' {
		panic(fmt.Sprintf("web: 路由冲突, 重复注册[%s]", path))
	}

	// 根节点特殊处理下
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突, 重复注册[/]")
		}
		root.handler = handleFunc
		return
	}

	// 结尾的检验
	if path[len(path)-1] == '/' {
		panic("web: 路径不能以 / 结尾")
	}

	// 中间连续 //, 可以用 strings.Contains()

	// 切割这个 path
	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == "" {
			panic("web: 不能有连续的 /")
		}

		// 递归下去, 找准位置
		// 如果中途有节点不存在, 你就要创建出来
		child := root.ChildOrCreate(seg)
		root = child
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突, 重复注册[%s]", path))
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*node, bool) {
	// 基本上是不是也是沿着树深度查找下去
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	// 根节点特殊处理
	if path == "/" {
		return root, true
	}

	// 这里把前置和后置的 / 都去掉
	path = strings.Trim(path, "/")
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		child, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		root = child
	}
	// 代表我确实有这个节点
	// 但是节点是不是用户注册的有 handler 的, 就不一定了
	return root, true
	//return root, root.handler != nil
}

type node struct {
	path string

	// 子 path 到字节点的映射
	// 静态节点
	children map[string]*node

	// 加一个通配符匹配
	starChild *node

	// 缺一个代表用户注册的业务逻辑
	handler HandleFunc
}

// 返回值是正确的子节点
func (n *node) ChildOrCreate(seg string) *node {
	if seg == "*" {
		n.starChild = &node{
			path: seg,
		}
		return n.starChild
	}

	if n.children == nil {
		n.children = map[string]*node{}
	}

	res, ok := n.children[seg]
	if !ok {
		// 如果不存在, 要新建
		res = &node{
			path: seg,
		}
		n.children[seg] = res
	}

	return res
}

// childOf 优先考虑静态匹配, 匹配不上再考虑通配符匹配
func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.starChild, n.starChild != nil
	}

	child, ok := n.children[path]
	if !ok {
		return n.starChild, n.starChild != nil
	}
	return child, ok
}
