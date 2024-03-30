package interceptor

import (
	"context"

	"connectrpc.com/connect"
)

// WithRedirector set up the context
func WithRedirector() connect.Option {
	type Location interface {
		GetLocationUri() string
	}

	interFn := func(next connect.UnaryFunc) connect.UnaryFunc {
		// prepare the callback
		fn := func(ctx context.Context, r connect.AnyRequest) (connect.AnyResponse, error) {
			// execute the method
			response, err := next(ctx, r)
			// set the location header
			if location, ok := response.Any().(Location); ok {
				response.Header().Set("X-Location", location.GetLocationUri())
			}

			return response, err
		}

		return fn
	}

	unaryFn := connect.UnaryInterceptorFunc(interFn)
	// prepare the option
	return connect.WithInterceptors(unaryFn)
}
