package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ralch/slogr"
)

// WithRecover recovers the handler from any panic.
func WithRecover() connect.HandlerOption {
	err := fmt.Errorf("the system is not in a state required for the operation's execution")

	return connect.WithRecover(
		func(ctx context.Context, _ connect.Spec, _ http.Header, r any) error {
			failure := fmt.Sprintf("%v", r)
			// prepare the logger
			logger := slogr.FromContext(ctx)
			logger.ErrorContext(ctx, "the system has an unexpected failure", slog.String("failure", failure))

			// return the error
			return connect.NewError(connect.CodeInternal, err)
		},
	)
}
