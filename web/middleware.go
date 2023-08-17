package web

/*
Middleware 函数式的责任链模式
或者叫做函数式的洋葱模式
*/
type Middleware func(next HandleFunc) HandleFunc

// AOP 方案在不同的框架, 不同的语言都有不同的叫法
// Middleware, Handler, Chain, Filter, Filter-Chain
/*
拦截器模式的设计
*/
//type MiddlewareV1 interface {
//	Invoke(next HandleFunc) HandleFunc
//}
//type Interceptor interface {
//	Before(ctx *Context)
//	After(ctx *Context)
//	Surround(ctx *Context)
//}

/*
集中式的设计,类似 gin
*/
//type Chain []HandleFuncV1
//
//type HandleFuncV1 func(ctx *Context) (next bool)
//type ChainV1 struct {
//	handler []HandleFuncV1
//}
//
//func (c ChainV1) Run(ctx *Context) {
//	for _, h := range c.handler {
//		next := h(ctx)
//		// 这种是中断执行
//		if !next {
//			return
//		}
//	}
//}

//type Net struct {
//	handlers []HandleFuncV1
//}
//
//func (n Net) Run(ctx *Context) {
//	var wg sync.WaitGroup
//	for _, hdl := range n.handlers {
//		h := hdl
//		if h.concurrent {
//			wg.Add(1)
//			go func() {
//				h.Run(ctx)
//				wg.Done()
//			}()
//		} else {
//			h.Run(ctx)
//		}
//	}
//	wg.Wait()
//}
//
//type HandleFuncV1 struct {
//	concurrent bool
//	handlers   []*HandleFuncV1
//}
//
//func (h HandleFuncV1) Run(ctx *Context) {
//	//for _, hdl := range h.handlers {
//	//	h := hdl
//	//	if h.concurrent {
//	//		wg.Add(1)
//	//		go func() {
//	//			h.Run(ctx)
//	//			wg.Done()
//	//		}()
//	//	}
//	//}
//}
