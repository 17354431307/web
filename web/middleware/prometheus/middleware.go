package prometheus

import (
	"strconv"
	"time"

	"github.com/Moty1999/web/web"
	"github.com/prometheus/client_golang/prometheus"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func (m MiddlewareBuilder) Build() web.Middleware {

	// 生成一个折线向量
	vec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      m.Name,
		Subsystem: m.Subsystem,
		Namespace: m.Namespace,
		Help:      m.Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"})

	// 注册到 prometheus
	prometheus.MustRegister(vec)
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime).Milliseconds()

				pattern := ctx.MatchedRoute
				if pattern == "" {
					pattern = "unknown"
				}

				// 响应时间
				vec.WithLabelValues(pattern, ctx.Req.Method, strconv.Itoa(ctx.RespStatusCode)).Observe(float64(duration))
			}()
			next(ctx)

		}
	}
}
