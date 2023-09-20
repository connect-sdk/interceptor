package interceptor

import (
	"context"

	connect "connectrpc.com/connect"
	idtoken "google.golang.org/api/idtoken"
	option "google.golang.org/api/option"
)

// WithAuthorization enables an oidc authorization.
func WithAuthorization(audience string, options ...option.ClientOption) connect.Option {
	innerFn := func(next connect.UnaryFunc) connect.UnaryFunc {
		fn := func(ctx context.Context, r connect.AnyRequest) (connect.AnyResponse, error) {
			if r.Spec().IsClient {
				// 1. prepare the source
				source, err := idtoken.NewTokenSource(ctx, audience, options...)
				if err != nil {
					return nil, connect.NewError(connect.CodeUnauthenticated, err)
				}
				// 2. prepare the token
				token, err := source.Token()
				if err != nil {
					return nil, connect.NewError(connect.CodeUnauthenticated, err)
				}

				// 3. prepare the header
				header := token.Type() + " " + token.AccessToken
				// 4. set the header
				r.Header().Set("Authorization", header)
			}

			return next(ctx, r)
		}

		return fn
	}

	unaryFn := connect.UnaryInterceptorFunc(innerFn)
	// prepare the option
	return connect.WithInterceptors(unaryFn)
}
