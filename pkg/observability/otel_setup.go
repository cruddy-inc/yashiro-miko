package observability

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func SetupLogger(ctx context.Context, serviceName string) (*log.LoggerProvider, error) {
	resources, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion("1.0.0"),
	))

	exporter, err := otlploghttp.New(ctx, otlploghttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	processor := log.NewBatchProcessor(exporter)
	return log.NewLoggerProvider(log.WithProcessor(processor), log.WithResource(resources)), nil
}

func SetupTrace(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}

	processor := sdktrace.NewBatchSpanProcessor(exporter)
	return sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(processor)), nil
}
