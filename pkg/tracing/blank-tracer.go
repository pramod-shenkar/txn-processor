package tracing

import "context"

type blankTracer struct{}
type blankSpan struct{}

// NewBlankTracer creates a no-op tracer
func NewBlankTracer() Tracer {
	return &blankTracer{}
}

func (t *blankTracer) Start(ctx context.Context, spanName string, opts ...any) (context.Context, Span) {
	return ctx, &blankSpan{}
}

func (t *blankTracer) Shutdown(ctx context.Context) error {
	return nil
}

func (s *blankTracer) IsEnabled() bool {
	return false
}

func (s *blankSpan) End()                  {}
func (s *blankSpan) RecordError(err error) {}
func (s *blankSpan) SetAttributes(...any)  {}
