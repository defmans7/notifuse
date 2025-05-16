package tracing

import (
	"context"
	"net/http"

	"go.opencensus.io/trace"
)

//go:generate mockgen -destination=../mocks/mock_tracer.go -package=pkgmocks github.com/Notifuse/notifuse/pkg/tracing Tracer

// Tracer defines the interface for tracing functionality
// codecov:ignore:start
type Tracer interface {
	// StartSpan starts a new span
	StartSpan(ctx context.Context, name string) (context.Context, *trace.Span)

	// StartSpanWithAttributes starts a new span with attributes
	StartSpanWithAttributes(ctx context.Context, name string, attrs ...trace.Attribute) (context.Context, *trace.Span)

	// StartServiceSpan starts a new span for a service method
	StartServiceSpan(ctx context.Context, serviceName, methodName string) (context.Context, *trace.Span)

	// EndSpan ends a span and records any error
	EndSpan(span *trace.Span, err error)

	// AddAttribute adds an attribute to the current span
	AddAttribute(ctx context.Context, key string, value interface{})

	// MarkSpanError marks the current span as failed with the given error
	MarkSpanError(ctx context.Context, err error)

	// TraceMethod is a helper to trace a service method with automatic span ending
	TraceMethod(ctx context.Context, serviceName, methodName string, f func(context.Context) error) error

	// TraceMethodWithResult is a helper to trace a service method that returns a result
	// Note: Due to Go interface limitations, we use interface{} instead of generics
	TraceMethodWithResultAny(ctx context.Context, serviceName, methodName string, f func(context.Context) (interface{}, error)) (interface{}, error)

	// WrapHTTPClient wraps an http.Client with OpenCensus tracing
	WrapHTTPClient(client *http.Client) *http.Client
}

// DefaultTracer is the default implementation of the Tracer interface
type DefaultTracer struct{}

// NewTracer creates a new DefaultTracer
func NewTracer() Tracer {
	return &DefaultTracer{}
}

// StartSpan implements Tracer.StartSpan
func (t *DefaultTracer) StartSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	return StartSpan(ctx, name)
}

// StartSpanWithAttributes implements Tracer.StartSpanWithAttributes
func (t *DefaultTracer) StartSpanWithAttributes(ctx context.Context, name string, attrs ...trace.Attribute) (context.Context, *trace.Span) {
	return StartSpanWithAttributes(ctx, name, attrs...)
}

// StartServiceSpan implements Tracer.StartServiceSpan
func (t *DefaultTracer) StartServiceSpan(ctx context.Context, serviceName, methodName string) (context.Context, *trace.Span) {
	return StartServiceSpan(ctx, serviceName, methodName)
}

// EndSpan implements Tracer.EndSpan
func (t *DefaultTracer) EndSpan(span *trace.Span, err error) {
	EndSpan(span, err)
}

// AddAttribute implements Tracer.AddAttribute
func (t *DefaultTracer) AddAttribute(ctx context.Context, key string, value interface{}) {
	AddAttribute(ctx, key, value)
}

// MarkSpanError implements Tracer.MarkSpanError
func (t *DefaultTracer) MarkSpanError(ctx context.Context, err error) {
	MarkSpanError(ctx, err)
}

// TraceMethod implements Tracer.TraceMethod
func (t *DefaultTracer) TraceMethod(ctx context.Context, serviceName, methodName string, f func(context.Context) error) error {
	return TraceMethod(ctx, serviceName, methodName, f)
}

// TraceMethodWithResultAny implements Tracer.TraceMethodWithResultAny
func (t *DefaultTracer) TraceMethodWithResultAny(ctx context.Context, serviceName, methodName string, f func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := StartServiceSpan(ctx, serviceName, methodName)
	defer span.End()

	result, err := f(ctx)
	if err != nil {
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: err.Error(),
		})
	}

	return result, err
}

// WrapHTTPClient implements Tracer.WrapHTTPClient
func (t *DefaultTracer) WrapHTTPClient(client *http.Client) *http.Client {
	return WrapHTTPClient(client)
}

// Global instance of the tracer
var globalTracer Tracer = NewTracer()

// GetTracer returns the global tracer instance
func GetTracer() Tracer {
	return globalTracer
}

// SetTracer sets the global tracer instance
func SetTracer(tracer Tracer) {
	globalTracer = tracer
}

// codecov:ignore:end
