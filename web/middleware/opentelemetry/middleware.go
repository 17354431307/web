package opentelemetry

import (
	"github.com/Moty1999/web/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/Moty1999/web/middleware/opentelemetry"
)

type MiddleWareBuild struct {
	Tracer trace.Tracer
}

//func NewMiddleWareBuild(tracer trace.Tracer) *MiddleWareBuild {
//	return &MiddleWareBuild{Tracer: tracer}
//}

func (m MiddleWareBuild) Build() web.Middleware {
	if m.Tracer == nil {
		m.Tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {

			// 尝试和客户端的 trace 结合在一起
			reqCtx := ctx.Req.Context()
			reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))

			reqCtx, span := m.Tracer.Start(reqCtx, "unknow")
			defer func() {
				// 这个执行完 next 才有值
				span.SetName(ctx.MatchedRoute)

				// 把响应码加上
				span.SetAttributes(attribute.Int("http.status", ctx.RespStatusCode))
				span.End()
			}()

			span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
			span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
			span.SetAttributes(attribute.String("http.schema", ctx.Req.URL.Scheme))
			span.SetAttributes(attribute.String("http.host", ctx.Req.Host))

			// 你这里还可以继续加

			ctx.Req = ctx.Req.WithContext(reqCtx)

			// 直接调用下一步
			next(ctx)
		}
	}
}
