package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type otelTracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

type otelSpan struct {
	span trace.Span
}

// NewOtelTracer creates a new OpenTelemetry tracer
func NewOtelTracer(ctx context.Context, serviceName, endpoint string) (Tracer, error) {

	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(endpoint),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)

	return &otelTracer{
		provider: tp,
		tracer:   otel.Tracer(serviceName),
	}, nil
}

func (t *otelTracer) Start(ctx context.Context, spanName string, opts ...any) (context.Context, Span) {
	ctx, span := t.tracer.Start(ctx, spanName)
	return ctx, &otelSpan{span: span}
}

func (t *otelTracer) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}

func (s *otelTracer) IsEnabled() bool {
	return true
}

func (s *otelSpan) End() {
	s.span.End()
}

func (s *otelSpan) RecordError(err error) {
	s.span.RecordError(err)
}

func (s *otelSpan) SetAttributes(attributes ...any) {
	attrs := make([]attribute.KeyValue, 0, len(attributes)/2)
	for i := 0; i < len(attributes)-1; i += 2 {
		key, ok := attributes[i].(string)
		if !ok {
			continue
		}
		switch v := attributes[i+1].(type) {
		case string:
			attrs = append(attrs, attribute.String(key, v))
		case int:
			attrs = append(attrs, attribute.Int(key, v))
		case int64:
			attrs = append(attrs, attribute.Int64(key, v))
		case bool:
			attrs = append(attrs, attribute.Bool(key, v))
		}
	}
	s.span.SetAttributes(attrs...)
}
