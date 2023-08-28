package test

import (
	"fmt"
	"github.com/Moty1999/web/web"
	"github.com/Moty1999/web/web/session"
	"github.com/Moty1999/web/web/session/cookie"
	"github.com/Moty1999/web/web/session/memory"
	"net/http"
	"testing"
	"time"
)

func TestSession(t *testing.T) {

	// 非常简单的登录验证

	var m *session.Manager = &session.Manager{
		Propagator: cookie.NewPropagator(),
		Store:      memory.NewStore(15 * time.Minute),
		CtxSessKey: "sessKey",
	}

	server := web.NewHTTPServer(web.ServerWithMiddleware(func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			if ctx.Req.URL.Path == "/login" {
				next(ctx)
				return
			}

			_, err := m.GetSession(ctx)
			if err != nil {
				ctx.RespStatusCode = http.StatusUnauthorized
				ctx.RespData = []byte("请重新登录")
				return
			}

			// 刷新 session 过期时间
			_ = m.RefreshSession(ctx)
			next(ctx)
		}
	}))

	server.Post("/login", func(ctx *web.Context) {
		// 要在这之前检验用户名和密码
		sess, err := m.InitSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("登录失败了")
			return
		}
		err = sess.Set(ctx.Req.Context(), "nickname", "hexiaowen")
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("登录失败了")
			return
		}
		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte("登录成功")
	})

	// 退出登录
	server.Post("/logout", func(ctx *web.Context) {
		// 清理各种数据
		err := m.RemoveSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("退出失败")
			return
		}
		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = []byte("退出登录")
	})

	server.Get("/user", func(ctx *web.Context) {
		sess, _ := m.GetSession(ctx)

		// 假如我要把昵称从 session 里面拿出来
		val, _ := sess.Get(ctx.Req.Context(), "nickname")
		fmt.Println(val)
		ctx.RespData = []byte(val.(string))
	})

	server.Start(":8081")
}
