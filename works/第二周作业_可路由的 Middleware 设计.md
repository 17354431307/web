# 可路由的 Middleware 设计



## 作业

在我们的课程里, 我们实现了非常简单的 server 级别 `middleware` 机制, 也就是 AOP 方案. 但是在实际问题中它虽然能够解决绝大部分问题, 但是在一些特定的问题上, 用起来并不是很方便. 



**现在的任务就是让我们设计的 Web 框架支持路由级别的 Middleware.** 



现在要支持新的功能特性, 就是允许用户指定某个 middleware 只对一些特定路径生效. 例如说: 用户注册
```go
Use("GET", "/a/b", ms)
```
那么只有 `/a/b` 这条路径能够执行 `ms` (`ms` 是注册的 `middleware`).



注: 这一次的作业我们并没有合并前面的路由作业, 所以我们还是在不支持正则路由和末尾通配符全匹配的基础上实现这个功能, 所以最终在这种场景之下:

```
Use("GET", "/a/*", ms)
```

如果输入的路径是 `/a/b/c/d/e` 这种, 虽然能够命中, 但是因为我们查找的时候返回 `mi,false`,  所以实际上没什么用.

预估完成时间: 3 小时



## 需求分析

### 场景分析

用户希望指定命中了某个路由的才能执行执行注册的 middleware. 假设我们的注册 Middleware 的 API 定义为:

```go
func (s *HTTPServer) Use(method string, path string, ms ...Middleware) {}
```



| 方法                                                         | 说明                                                         |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| `Use("GET", "/a/b", ms)`                                     | 当输入路径 `/a/b` 的时候, 会调度对应的 ms<br />当输入路径 `/a/b/c` 的时候，会会调度执行 ms |
| `Use("GET", "/a/*", ms1)`<br />`Use("GET", "/a/b/*", ms2)`   | 当输入路径 `/a/c` 的时候, 会调度 ms1<br/>当输入路径为 `/a/b/c` 的时候, 会调度 ms1 和 ms2 |
| `Use("GET", "/a/*/c", ms1)`<br />`Use("GET", "/a/b/c", ms2)` | 当输入路径 `/a/d/c` 的时候，调度执行 ms1<br/>当输入路径 `/a/b/c` 的时候，调度执行 ms1 和 ms2<br/>当输入路径 `/a/b/d` 不会调度执行 ms1 或者 ms2 |
| `Use("GET", "/a/:id", ms1)`<br />`Use("GET", "/a/123/c", ms2)` | 当输入路径 `/a/123` 的时候，调度执行 ms1<br/>当输入路径 `/a/123/c` 的时候。调度执行 ms1 和 ms2 |

然而这一类场景:

| `Use("GET", "/a/:id", ms1)`<br />`Use("GET", "/a/123/c", ms2)` | 当输入路径 `/a/b/c` 的时候，调度执行 ms2 |
| ------------------------------------------------------------ | ---------------------------------------- |

也就是，精确命中的路由上注册的 ms 才允许被调度。这一类场景在实际中很少使用，大部分
都是希望同时调度 ms1和 ms2。

例如在实际中，用户注册两个：

```go
Use("GET", "/*", AccessLog)
Use("GET", "/a/b/c", Auth)
```

很显然，用户是希望在输入路径 `/a/b/c` 的时候既能够输出访问日志，也能够执行鉴权。

也就是，在查找路由的时候，我们选择的是最精确匹配；**那么在查找 middleware 的时候，我们执行的是能匹配则匹配原则。**    

另外一个问题是，在能够匹配上多个的情况下，就要考虑它们之间的先后顺序。在这个时候，我们**遵循的原则是，越具体越后调度**。

```go
Use("GET", "/a/b", ms1)
Use("GET", "/a/*", ms2)
Use("GET", "/a", ms3)
```

那么我们的执行顺序是 ms3, ms2, ms1．

### 功能需求

+ 允许用户在特定的路由上注册 `middleware`
+ middleware 选取所有能够匹配上的路由的 `middleware` 作为结果



### 非功能需求

+ 对路由树性能影响有限（这个影响阈值，要通过基准测试来确定）



## 设计

在场景分析的时候我们总结到，middleware 的选取是能取尽取。实际上，这意味着我们要拿着路径的每一段，然后在整颗路由树里面进行查找，而且在查找子节点的时候，不仅仅是找最精确匹配的那个，而是找所有能够匹配上的子节点。而后以这些子节点作为起点，进一步深入查找。

这些找到的节点（包含中间过程的节点），只要有 middleware，就是我们需要的 middleware。

假如说我们的路由树是：

![image-20230818165609015](.\assets\image-20230818165609015.png)

假如说我们要查找 `/a/b/c` 命中的 `middleware`，那么按照我们的原则：

+ 首先遍历第一层，只有 `a` 一个节点，那么我们将 `a` 的 middleware 加入到最终结果里面.
+ 以 `a`  的子节点作为候选节点，它们是第二层。依次遍历，那么能够命中 `b` 的就有 `b` 和 `*` 两个节点，将这两个节点的 middleware 加入到最终结果里面.
+ 以 `b` 和 `*` 的子节点作为第三层，只有一个 `*`。将它的 middleware 加入到最终结果集里面
+ 结束，返回最终结果

所以我们可以清晰看到，这里我们是利用了层次遍历来找到最终的 middleware.

**提示: 其实这就是一个多叉树的广度优先遍历, 可以利用队列结构来进行层级遍历, `GO`** 语言标准库有实现的队列 `container/list`, 它是一个双端的队列, 不过不支持泛型, 用起来不那么优雅, 现在 `GO`  团队已经在用泛型改写标准库, 静等支持.

### 详细设计

API定义为:

```go
// Use 会执行路由匹配, 只有匹配上了的 mdls 才会生效
// 这个只需要稍微改造一下路由树就可以实现
func (s *HTTPServer) Use(method string, path string, mdls ...Middleware) (*matchInfo, bool) {
    panic("implement me!")
    s.addRoute(method, path, nil, mdls...)
}
```

这里我们为 `addRoute` 加了一个新的参数 `mds...`，也就是我们直接依托原本的路由树来完成这个功能

```go
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
    root, ok := r.trees[method]
    if !ok {
        return nil, false
    }
    
    if path == "/" {
        return &matchInfo{n: root, mdls: root.mdls}, true
    }
    
    segs := strings.split(strings.Trim(path, "/"), "/")
    mi := &matchInfo{}
    cur := root
    for _, s := range segs {
        var matchParam bool
        cur, matchParam ok = cur.childOf(s)
        if matchParam {
            mi.addValue(root.path[1:], s)
        }
    }
    mi.n = cur
    mi.mdls = r.findMdls(root, segs)
    return mi, true
}

func (r *router) findMdls(root *node, segs []string) []Middleware {
    // 层次遍历
    pinic("implement me!")
}

```



## 可替换方案

### 提前计算

注意到，在这里我们每次路由匹配的时候都是重新算了一下 `mi.mds`。实际上这个部分有两种优化手段：

+ 提前计算：在服务器启动的时候提前计算好所有叶子节点的 mdls，也就是提前按照层次遍历，为每一个有 `handler` 的节点提前计算好 mdls。路由查找的时候就可以避免重复遍历.
  ```go
  // calculateRouteMdls 计算路由的 midls
  func (r *router) calculateRouteMdls() {
  	var fn func(root *route)
  	fn = func(root *route) {
  		queue := make([]*route, 0, 20)
  		queue = append(queue, root)
  
  		for len(queue) > 0 {
  			nextQueue := []*route{}
  			for _, cur := range queue {
  				mdls := []Middleware{}
  				if cur.anyChild != nil {
  					mdls = append(mdls, cur.anyChild.mdls...)
  
  					tempSlice := make([]Middleware, 0, len(cur.anyChild.mdls)+len(cur.mdls))
  					tempSlice = append(tempSlice, cur.mdls...)
  					tempSlice = append(tempSlice, cur.anyChild.mdls...)
  					cur.anyChild.mdls = tempSlice
  
  					nextQueue = append(nextQueue, cur.anyChild)
  				}
  
  				if cur.paramChild != nil {
  					mdls = append(mdls, cur.paramChild.mdls...)
  
  					tempSlice := make([]Middleware, 0, len(cur.paramChild.mdls)+len(cur.mdls))
  					tempSlice = append(tempSlice, cur.mdls...)
  					tempSlice = append(tempSlice, cur.paramChild.mdls...)
  					cur.paramChild.mdls = tempSlice
  
  					nextQueue = append(nextQueue, cur.paramChild)
  				}
  
  				var regMiddlewares []Middleware
  				if cur.regChild != nil {
  					regMiddlewares = cur.regChild.mdls
  					// 这个就比较特殊了
  					// 因为静态子节点必须满足这个正则才能加上去
  
  					tempSlice := make([]Middleware, 0, len(cur.regChild.mdls)+len(cur.mdls))
  					tempSlice = append(tempSlice, cur.mdls...)
  					tempSlice = append(tempSlice, cur.regChild.mdls...)
  					cur.regChild.mdls = tempSlice
  
  					nextQueue = append(nextQueue, cur.regChild)
  				}
  
  				for _, child := range cur.children {
  
  					// 判断子节点是否满足 regChild
  					if cur.regChild != nil && cur.regChild.regExec != nil && cur.regChild.regExec.Match([]byte(child.path)) {
  						mdls = append(mdls, regMiddlewares...)
  					}
  
  					tempSlice := make([]Middleware, 0, len(cur.mdls)+len(mdls)+len(child.mdls))
  					tempSlice = append(tempSlice, cur.mdls...)
  					tempSlice = append(tempSlice, mdls...)
  					tempSlice = append(tempSlice, child.mdls...)
  
  					child.mdls = tempSlice
  					nextQueue = append(nextQueue, child)
  				}
  			}
  			queue = nextQueue
  		}
  
  	}
  
  	for _, root := range r.trees {
  		fn(root)
  	}
  }
  ```

  
+ 避免重复计算。第一次计算好之后，我们将结果保存下来。但是因为我们路由匹配原本是一个纯粹的查找过程，所以我们不需要使用并发保护。但是现在我们成了一个读写过程，那么就需要考虑使用 `sync.Once` 或者 `atomic` 来完成读写操作

### 简化方案

这里我们的场景是支持了最为丰富的场景。也就是说和路由匹配的逻辑是一模一样，只是说路由匹配是找最精确的，而 middleware是找所有能匹配上的。

另外只支持这种场景：

| `Use("GET", "/a/*", ms1)`   | 用户输入 `/a` 不会执行 ms1 <br/>用户输入 `/a/b` 会执行 ms1<br/>用户输入 `/a/b/c` 会执行 ms1 |
| --------------------------- | ------------------------------------------------------------ |
| `Use("GET", "/a/:id", ms1)` | 禁止注册 middleware 的时候使用路径参数，<br/>类似地，也禁止正则匹配 |
| `Use("GET", "/a/b", ms1)`   | 这种可以禁止，也可以不禁止。<br />如果不禁止，那么只有 `/a/b` 的时候才会执行 ms1 |
| `Use("GET", "/a/*/c", ms1)` | 这种禁止。                                                   |

这种支持起来代码更加简单。只需要将查找路径上的静态节点的 middleware 合并在一起就可以



## 测试

### 单元测试

单元测试要考虑：

+ 静态匹配，通配符匹配和正则匹配三个的组合情况
+ 要考虑注册和 `/` 相关的情况，因为我们的代码是对 `/` 有一些特殊的处理
+ 要考虑 middleware 的顺序

基准测试

综合比较在没有引入该功能，还引入该功能后的路由匹配性能



## 答案

自己写的:

```go
func (r *router) addRoute(method string, path string, handleFunc HandleFunc, mdls ...Middleware) {
    // ...
    // 根节点特殊处理下
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突, 重复注册[/]")
		}
		root.handler = handleFunc
		root.mdls = mdls
		root.route = "/"
		return
	}
    
    
    //...
    root.handler = handleFunc
	root.mdls = mdls
	root.route = path
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
    // ...
    
    // 根节点特殊处理
	if path == "/" {
		return &matchInfo{node: root, mdls: root.mdls}, true
	}
    
    // ...
    mi.mdls = r.findMdls(root, segs)
	mi.node = cur
	return mi, true
}

func (r *router) findMdls(root *node, segs []string) []Middleware {

	mdls := []Middleware{}

	queue := list.New()
	queue.PushBack(root)
	for _, seg := range segs {
		// 一层一层的处理
		l := queue.Len()
		for l > 0 {
			element := queue.Front()
			queue.Remove(element)
			node := element.Value.(*node)
			if node.mdls != nil {
				mdls = append(mdls, node.mdls...)
			}

			// 先从通配符匹配
			if node.starChild != nil {
				queue.PushBack(node.starChild)
			}

			// 静态节点找
			child, ok := node.children[seg]
			if ok {
				queue.PushBack(child)
			}

			l--
		}
	}

	if queue.Len() > 0 {
		for element := queue.Front(); element != nil; element = element.Next() {
			node := element.Value.(*node)
			if node.mdls != nil {
				mdls = append(mdls, node.mdls...)
			}
		}
	}
	return mdls
}

func (h *HTTPServer) serve(ctx *Context) {
	// before route

	// 接下来查找路由并且执行命中的业务逻辑
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	// after route

	var root HandleFunc = func(ctx *Context) {
		if !ok || info.node.handler == nil {
			// 路由没有命中
			ctx.RespStatusCode = 404
			ctx.RespData = []byte("Not Found")

			return
		}

		ctx.PathParams = info.pathParams
		ctx.MatchedRoute = info.node.route
		// before execute
		info.node.handler(ctx)
		// after execute
	}

	// 构建路由级别的中间件
	mdls := info.mdls
	for i := len(mdls) - 1; i >= 0; i-- {
		root = mdls[i](root)
	}
	root(ctx)
}
```

老师给的:

```go
func (r *router) findMdls(root *node, segs []string) []Middleware {
	queue := []*node{root}
	res := make([]Middleware, 0, 16)
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		var children []*node
		for _, cur := range queue {
			if len(cur.mdls) > 0 {
				res = append(res, cur.mdls...)
			}

			children = append(children, cur.childrenOf(seg)...)
		}
		queue = children
	}

	for _, cur := range queue {
		if len(cur.mdls) > 0 {
			res = append(res, cur.mdls...)
		}
	}
	return res
}

func (n *node) childrenOf(path string) []*node {
	res := make([]*node, 0, 4)
	var static *node
	if n.children != nil {
		static = n.children[path]
	}
	if n.starChild != nil {
		res = append(res, n.starChild)
	}
	if n.paramChild != nil {
		res = append(res, n.paramChild)
	}
	if static != nil {
		res = append(res, static)
	}
	return res
}


```

