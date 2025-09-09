package obs

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

type Shutdown func(context.Context) error

type Setup struct {
	Stop Shutdown
}

func Init(ctx context.Context, endpoint, token string) (*Setup, error) {
	if endpoint == "" || token == "" {
		slog.Warn("Endpoint or token is empty")
		return &Setup{Stop: func(context.Context) error { return nil }}, nil
	}

	exp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithURLPath("/otlp/v1/metrics"),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + token,
		}),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("wod-gen"),
			semconv.ServiceVersionKey.String("v1"),
			semconv.DeploymentEnvironmentKey.String(os.Getenv("ENV")),
		),
	)
	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(10*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return &Setup{Stop: mp.Shutdown}, nil
}
