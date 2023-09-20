package interceptor

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid"
	"github.com/ralch/slogr"
)

// WithLogger set up the logger.
func WithLogger() connect.Option {
	interFn := func(next connect.UnaryFunc) connect.UnaryFunc {
		// prepare the callback
		fn := func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			logger := slogr.FromContext(ctx)

			var (
				id   = uuid.Must(uuid.NewV4()).String()
				spec = request.Spec()
				peer = request.Peer()
			)

			uri := &url.URL{
				Scheme: peer.Protocol,
				Host:   peer.Addr,
				Path:   spec.Procedure,
			}

			message := fmt.Sprintf("[%v] %v", spec.StreamType, uri)

			// log the start
			logger.InfoContext(ctx, message,
				slogr.OperationStart(id, spec.Procedure),
			)

			// prepare the context
			ctx = slogr.WithContext(ctx, logger.With(
				slogr.OperationContinue(id, spec.Procedure),
			))

			// execute the method
			response, err := next(ctx, request)
			if err == nil {
				// log the end
				logger.InfoContext(ctx, message,
					slogr.OperationEnd(id, spec.Procedure),
				)
			} else {
				// log the end
				logger.ErrorContext(ctx, message,
					slogr.OperationEnd(id, spec.Procedure),
					slogr.Error(err),
				)
			}

			return response, err
		}
		// done!
		return fn
	}

	unaryFn := connect.UnaryInterceptorFunc(interFn)
	// prepare the option
	return connect.WithInterceptors(unaryFn)
}
