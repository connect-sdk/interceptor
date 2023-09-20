package middleware

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// WithTracer set up the Open Telemetry Tracer.
func WithTracer() connect.Interceptor {
	return otelconnect.NewInterceptor()
}

var _ propagation.TextMapPropagator = TracePropagator{}

// TracePropagator implements propagation.TracePropagator to propagate
// traces in HTTP headers for Google Cloud Platform and Stackdriver Trace.
type TracePropagator struct{}

// NewTracePropagator creates a new propagator.
func NewTracePropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		// TraceContext is a propagator that supports the W3C Trace Context format.
		propagation.TraceContext{},
		// Baggage is a propagator that supports the W3C Baggage format.
		propagation.Baggage{},
		// TracePropagator is a propagator that support the Google Cloud Platform format.
		TracePropagator{},
	)
}

// Inject injects the span into a given carrier.
func (f TracePropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()

	if !spanCtx.TraceID().IsValid() || !spanCtx.SpanID().IsValid() {
		return
	}

	// prepare the span id
	var (
		id        = spanCtx.SpanID()
		idEncoded = binary.BigEndian.Uint64(id[:])
	)
	// prepare the value
	value := fmt.Sprintf("%s/%d;o=%d", spanCtx.TraceID().String(), idEncoded, spanCtx.TraceFlags())
	// setht the value
	carrier.Set("X-Cloud-Trace-Context", value)
}

// Extract returns the context from a given carrier.
func (f TracePropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	if value := carrier.Get("X-Cloud-Trace-Context"); value != "" {
		spanCtx, err := f.extract(value)
		if err == nil && spanCtx.IsValid() {
			return trace.ContextWithRemoteSpanContext(ctx, spanCtx)
		}
	}

	return ctx
}

// extract is using functionality from https://github.com/glassonion1/logz/blob/main/propagation/http_format.go
func (f TracePropagator) extract(header string) (trace.SpanContext, error) {
	spanCtx := trace.SpanContext{}

	if header == "" || len(header) > 200 {
		return trace.SpanContext{}, fmt.Errorf("header over max size")
	}

	// Parse the trace id field.
	slash := strings.Index(header, `/`)
	if slash == -1 {
		return spanCtx, errors.New("failed to parse value")
	}

	tid, header := header[:slash], header[slash+1:]
	traceID, err := trace.TraceIDFromHex(tid)
	if err != nil {
		return spanCtx, fmt.Errorf("failed to parse value: %w", err)
	}

	spanCtx = spanCtx.WithTraceID(traceID)
	// Parse the span id field.
	spanValue := header
	semicolon := strings.Index(header, `;`)
	if semicolon != -1 {
		spanValue, header = header[:semicolon], header[semicolon+1:]
	}

	sid, err := strconv.ParseUint(spanValue, 10, 64)
	if err != nil {
		return spanCtx, fmt.Errorf("failed to parse value: %w", err)
	}
	spanID := spanCtx.SpanID()
	binary.BigEndian.PutUint64(spanID[:], sid)
	spanCtx = spanCtx.WithSpanID(spanID)

	// Parse the options field, options field is optional.
	if !strings.HasPrefix(header, "o=") {
		return spanCtx, errors.New("failed to parse value")
	}

	value, err := strconv.ParseUint(header[2:], 10, 64)
	if err != nil {
		return spanCtx, fmt.Errorf("failed to parse value: %w", err)
	}

	// 1 = to sample
	if value == 1 {
		spanCtx = spanCtx.WithTraceFlags(trace.FlagsSampled)
	}

	return spanCtx, nil
}

func (f TracePropagator) Fields() []string {
	return []string{"X-Cloud-Trace-Context"}
}
