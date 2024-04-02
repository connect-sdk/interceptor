package interceptor

import (
	"context"
	"net/url"

	"connectrpc.com/connect"
	"google.golang.org/grpc/metadata"
)

// WithContext set up the context
func WithContext() connect.Option {
	interFn := func(next connect.UnaryFunc) connect.UnaryFunc {
		// prepare the callback
		fn := func(ctx context.Context, r connect.AnyRequest) (connect.AnyResponse, error) {
			if r.Spec().IsClient {
				kv, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					kv = metadata.MD{}
				}

				// copy the headers to the context
				for k, vv := range kv {
					for _, v := range vv {
						r.Header().Add(k, v)
					}
				}
			} else {
				kv, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					kv = metadata.MD{}
				}
				// copy the headers to the context
				for k, v := range r.Header() {
					kv.Set(k, v...)
				}
				// prepare the context
				ctx = metadata.NewIncomingContext(ctx, kv)
			}
			// execute the method
			return next(ctx, r)
		}

		return fn
	}

	unaryFn := connect.UnaryInterceptorFunc(interFn)
	// prepare the option
	return connect.WithInterceptors(unaryFn)
}

// GetRequestURL returns the request URL
func GetRequestURL(r connect.AnyRequest) *url.URL {
	uri := &url.URL{
		Scheme: "http",
		Host:   r.Header().Get("Host"),
	}

	if r.Header().Get("X-Forwarded-Host") != "" {
		uri.Host = r.Header().Get("X-Forwarded-Host")
	}

	if r.Header().Get("X-Forwarded-Proto") != "" {
		uri.Scheme = r.Header().Get("X-Forwarded-Proto")
	}

	return uri
}
