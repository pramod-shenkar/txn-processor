package tracing

import "context"

// Tracer is the interface for distributed tracing
type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...any) (context.Context, Span)
	Shutdown(ctx context.Context) error
	IsEnabled() bool
}

// Span represents a tracing span
type Span interface {
	End()
	RecordError(err error)
	SetAttributes(attributes ...any)
}
