//go:build e2e

package opentelemetry

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/Moty1999/web/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func TestMiddleWareBuild_Build(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	build := MiddleWareBuild{
		Tracer: tracer,
	}
	server := web.NewHTTPServer(web.ServerWithMiddleware(build.Build()))

	server.Get("/user", func(ctx *web.Context) {
		first1C, first1 := tracer.Start(ctx.Req.Context(), "first1_layer")
		defer first1.End()

		secondC, second := tracer.Start(first1C, "second_layer")
		time.Sleep(time.Second)
		_, third1 := tracer.Start(secondC, "third_layer_1")
		time.Sleep(100 * time.Millisecond)
		third1.End()

		_, third2 := tracer.Start(secondC, "third_layer_2")
		time.Sleep(300 * time.Millisecond)
		third2.End()
		second.End()

		_, first2 := tracer.Start(ctx.Req.Context(), "first_layer_2")
		defer first2.End()
		time.Sleep(100 * time.Millisecond)
		ctx.RespJSON(202, User{
			Name: "Tom",
		})
	})

	initZipkin(t)
	server.Start(":8081")
}

type User struct {
	Name string
}

func initZipkin(t *testing.T) {
	exporter, err := zipkin.New(
		"http://localhost:19411/api/v2/spans",
		zipkin.WithLogger(log.New(os.Stderr, "opentelemetry-demo", log.Ldate|log.Ltime|log.Llongfile)),
	)
	if err != nil {
		t.Fatal(err)
	}

	batcher := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("opentelemetry-demo"),
		)),
	)
	otel.SetTracerProvider(tp)
}

func initJeager(t *testing.T) {
	url := "http://localhost:14268/api/traces"
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		t.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		// Always be sure to batch in production.
		sdktrace.WithBatcher(exp),
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("opentelemetry-demo"),
			attribute.String("environment", "dev"),
			attribute.Int64("ID", 1),
		)),
	)

	otel.SetTracerProvider(tp)
}
