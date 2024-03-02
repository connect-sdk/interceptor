package interceptor

import (
	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/connect-sdk/telemetry"
)

// WithTracer set up the Open Telemetry Tracer.
func WithTracer() connect.Option {
	propagator := telemetry.NewTracePropagator()

	unaryFn, _ := otelconnect.NewInterceptor(
		otelconnect.WithTrustRemote(),
		otelconnect.WithPropagator(propagator),
	)

	return connect.WithInterceptors(unaryFn)
}
