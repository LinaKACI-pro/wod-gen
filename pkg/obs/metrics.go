package obs

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const Increment = 1

func MetricsMiddleware() gin.HandlerFunc {
	meter := otel.GetMeterProvider().Meter("wod-gen/http")
	reqCounter, _ := meter.Int64Counter("http_server_requests_total")
	latencyHist, _ := meter.Float64Histogram("http_server_request_duration_ms")

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		durMs := float64(time.Since(start).Milliseconds())
		attrs := []attribute.KeyValue{
			attribute.String("method", c.Request.Method),
			attribute.Int("status", c.Writer.Status()),
			attribute.String("route", c.FullPath()),
		}
		reqCounter.Add(c, Increment, metric.WithAttributes(attrs...))
		latencyHist.Record(c, durMs, metric.WithAttributes(attrs...))
	}
}
