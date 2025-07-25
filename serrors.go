package serrors

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// StructuredError represents a structured, optionally correlated error.
type StructuredError struct {
	Msg     string
	Op      string
	Err     error
	Fields  map[string]any
	TraceID string // Optional OTEL trace ID
	SpanID  string // Optional OTEL span ID
}

func (e *StructuredError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Msg
}

func (e *StructuredError) Unwrap() error {
	return e.Err
}

func (e *StructuredError) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, len(e.Fields)+3)

	for k, v := range e.Fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	attrs = append(attrs, slog.String("op", e.Op))
	if e.Err != nil {
		attrs = append(attrs, slog.String("cause", e.Err.Error()))
	}
	if e.TraceID != "" {
		attrs = append(attrs, slog.String("trace_id", e.TraceID))
	}
	if e.SpanID != "" {
		attrs = append(attrs, slog.String("span_id", e.SpanID))
	}

	return slog.GroupValue(attrs...)
}

// New creates a StructuredError with optional span correlation.
func New(ctx context.Context, op, msg string, err error, fields map[string]any) error {
	se := &StructuredError{
		Msg:    msg,
		Op:     op,
		Err:    err,
		Fields: fields,
	}

	if span := trace.SpanFromContext(ctx); span != nil && span.SpanContext().IsValid() {
		sc := span.SpanContext()
		se.TraceID = sc.TraceID().String()
		se.SpanID = sc.SpanID().String()
	}

	return se
}
