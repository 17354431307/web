package recover

import "github.com/Moty1999/web/web"

type MiddlewareBuilder struct {
	StatusCode int
	Data       []byte
	//log func(err any)
	Log func(ctx *web.Context)
	//log func(stack string)
}

func (m MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespData = m.Data
					ctx.RespStatusCode = m.StatusCode
					m.Log(ctx)
				}
			}()
			next(ctx)
		}
	}
}
