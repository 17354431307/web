package web

import (
	"fmt"
	"regexp"
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
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	// 基本上是不是也是沿着树深度查找下去
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	// 根节点特殊处理
	if path == "/" {
		return &matchInfo{node: root}, true
	}

	// 这里把前置和后置的 / 都去掉
	segs := strings.Split(strings.Trim(path, "/"), "/")
	mi := &matchInfo{}
	for _, seg := range segs {
		child, ok := root.childOf(seg)
		if !ok {
			if root.typ == nodeTypeAny {
				mi.node = root
				return mi, true
			}

			return nil, false
		}

		// 命中了路径参数
		if child.paramName != "" {
			mi.addValue(child.paramName, seg)
		}

		root = child
	}
	// 代表我确实有这个节点
	// 但是节点是不是用户注册的有 handler 的, 就不一定了

	mi.node = root
	return mi, true
}

type nodeType int

const (
	// 静态路由
	nodeTypeStatic = iota
	// 正则路由
	nodeTypeReg
	// 路径参数路由
	nodeTypeParam
	// 通配符匹配
	nodeTypeAny
)

type node struct {
	typ  nodeType
	path string
	// 缺一个代表用户注册的业务逻辑
	handler HandleFunc

	// 子 path 到字节点的映射
	// 静态节点
	children map[string]*node

	// 加一个通配符匹配
	starChild *node

	// 加一个路径参数
	paramsChild *node

	// 正则路由和参数路由都会使用这个字段
	paramName string

	// 正则表达式
	regChild *node
	regExpr  *regexp.Regexp
}

// 返回值是正确的子节点
func (n *node) ChildOrCreate(path string) *node {
	if path == "*" {
		if n.paramsChild != nil {
			panic("web: 非法路由, 已有路径参数路由. 不允许同时注册通配符路由和路径参数路由")
		}

		if n.regChild != nil {
			panic("web: 非法路由, 已有正则路由. 不允许同时注册通配符路由和正则路由")
		}

		if n.starChild == nil {
			n.starChild = &node{
				path: path,
				typ:  nodeTypeAny,
			}
		}
		return n.starChild
	}

	// 以 : 开头, 需要进一步解析, 判断是路径参数路由还是正则路由
	if path[0] == ':' {
		parseName, expr, isReg := n.parseParam(path)
		if isReg {
			n.childOrCreateReg(path, expr, parseName)
		}

		return n.childOrCreateParam(path, parseName)
	}

	if n.children == nil {
		n.children = map[string]*node{}
	}

	child, ok := n.children[path]
	if !ok {
		// 如果不存在, 要新建
		child = &node{
			path: path,
			typ:  nodeTypeStatic,
		}
		n.children[path] = child
	}

	return child
}

/*
childOf 优先考虑静态匹配, 匹配不上再考虑通配符匹配

	return:
		*node: 找到的节点
		bool: 是否命中
*/
func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.childOfNonStatic(path)
	}

	child, ok := n.children[path]
	if !ok {
		return n.childOfNonStatic(path)
	}
	return child, ok
}

// parseParam 用于解析是不是正则表达式
// 第一个返回值是参数名字
// 第二个返回值是正则表达式
// 第三个返回值为 true 则说明是正则路由
var reg = regexp.MustCompile(`:(\w+)(?:\((\w+)\))?`)

func (n *node) parseParam(path string) (string, string, bool) {

	ret := reg.FindStringSubmatch(path)[1:]

	if ret[1] != "" {
		return ret[0], ret[1], true
	}

	return ret[0], "", false
}

func (n *node) childOrCreateReg(path string, expr string, paramName string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.paramsChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.regChild != nil {
		if n.regChild.regExpr.String() != expr || n.paramName != paramName {
			panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("web: 正则表达式错误 %w", err))
		}
		n.regChild = &node{
			path:      path,
			paramName: paramName,
			regExpr:   regExpr,
			typ:       nodeTypeReg,
		}
	}

	return n.regChild
}

func (n *node) childOrCreateParam(path string, paramName string) *node {
	if n.regChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有正则路由. 不允许同时注册正则路由和参数路由 [%s]", path))
	}

	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
	}

	if n.paramsChild != nil {
		if n.paramsChild.paramName != paramName {
			panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramsChild.path, path))
		}
	} else {
		n.paramsChild = &node{
			typ:       nodeTypeParam,
			path:      path,
			paramName: paramName,
		}
	}

	return n.paramsChild
}

// childOfNonStatic 从非静态匹配的子节点里面找
func (n *node) childOfNonStatic(path string) (*node, bool) {

	if n.regChild != nil && n.regChild.regExpr.Match([]byte(path)) {
		return n.regChild, true
	}
	if n.paramsChild != nil {
		return n.paramsChild, true
	}

	return n.starChild, n.starChild != nil
}

type matchInfo struct {
	node       *node
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}
