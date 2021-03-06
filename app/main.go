package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	// Initialize tracer with a logger and a metrics factory
	t, closer := newTracer("apm-go-playground", "jaeger:5775")
	defer closer.Close()
	opentracing.SetGlobalTracer(t)

	// Init Gin
	g := gin.Default()

	// Use Tracing Middleware
	g.Use(openTracingMiddleware(opentracing.GlobalTracer()))

	g.GET("/status", status)

	// Start Server
	log.Fatalln("Start Server Error", g.Run("0.0.0.0:9000"))
}

func status(c *gin.Context) {

	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), "Check Deps Status")
	defer span.Finish()

	span.LogEvent("Pretending to check external deps")

	rSpan, _ := opentracing.StartSpanFromContext(ctx, "Check Redis Status")
	rSpan.SetTag("redis.status", "OK")
	time.Sleep(time.Millisecond * 10)
	rSpan.Finish()

	span.LogEvent("External deps ok")

	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
	return
}
func openTracingMiddleware(t opentracing.Tracer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//span := t.StartSpan(ctx.Request.RequestURI)
		span, tracingCtx := opentracing.StartSpanFromContext(ctx.Request.Context(), "New HTTP Request")
		span.SetTag("http.uri", ctx.Request.RequestURI)
		span.SetTag("http.method", ctx.Request.Method)
		span.SetTag("http.user_agent", ctx.Request.UserAgent())
		span.SetTag("http.client_ip", ctx.ClientIP())
		ctx.Request = ctx.Request.WithContext(tracingCtx)
		defer span.Finish()

		//ctx.Set("tracer",tracer.)
		ctx.Next()

		span.SetTag("http.status_code", ctx.Writer.Status())
		span.SetTag("http.response_size", ctx.Writer.Size())
	}
}

func newTracer(serviceName, hostPort string) (opentracing.Tracer, io.Closer) {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
			//CollectorEndpoint:   "http://localhost:14268",
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  hostPort, // localhost:5775
		},
	}
	tracer, closer, err := cfg.New(
		serviceName,
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		log.Fatal(err)
	}

	return tracer, closer
}
