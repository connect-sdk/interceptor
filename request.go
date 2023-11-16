package interceptor

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/grpc/metadata"
)

// WithIncomingContext set up the incoming context
func WithIncomingContext() connect.Option {
	interFn := func(next connect.UnaryFunc) connect.UnaryFunc {
		// prepare the callback
		fn := func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			kv, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				kv = metadata.MD{}
			}
			// copy the headers to the incoming context
			for k, v := range request.Header() {
				kv.Set(k, v...)
			}

			rctx := metadata.NewIncomingContext(ctx, kv)
			// execute the method
			return next(rctx, request)
		}

		return fn
	}

	unaryFn := connect.UnaryInterceptorFunc(interFn)
	// prepare the option
	return connect.WithInterceptors(unaryFn)
}

// NewOutgoingRequest wraps a generated request message.
func NewOutgoingRequest[T any](ctx context.Context, message *T) *connect.Request[T] {
	request := connect.NewRequest[T](message)

	if kv, ok := metadata.FromOutgoingContext(ctx); ok {
		header := request.Header()

		for k, values := range kv {
			for _, v := range values {
				header.Add(k, v)
			}
		}
	}

	return request
}
