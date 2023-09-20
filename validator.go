package interceptor

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// WithValidator set up the request validator
func WithValidator() connect.Option {
	type RequestWithDefault interface {
		Default()
	}

	type RequestWithValidation interface {
		ValidateAll() error
	}

	type Error interface {
		Field() string
		Reason() string
	}

	type ErrorCollection interface {
		AllErrors() []error
	}

	interFn := func(next connect.UnaryFunc) connect.UnaryFunc {
		// prepare the callback
		fn := func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			// prepare the default values
			if value, ok := request.Any().(RequestWithDefault); ok {
				value.Default()
			}

			if value, ok := request.Any().(RequestWithValidation); ok {
				if err := value.ValidateAll(); err != nil {
					switch xerr := err.(type) {
					case ErrorCollection:
						verr := &errdetails.BadRequest{}

						for _, item := range xerr.AllErrors() {
							if ferr, ok := item.(Error); ok {
								verr.FieldViolations = append(verr.FieldViolations,
									&errdetails.BadRequest_FieldViolation{
										Field:       ferr.Field(),
										Description: ferr.Reason(),
									},
								)
							}
						}

						detail, derr := connect.NewErrorDetail(verr)
						if derr != nil {
							return nil, derr
						}

						cerr := connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("bad request"))
						cerr.AddDetail(detail)

						return nil, cerr
					default:
						return nil, connect.NewError(connect.CodeInvalidArgument, err)
					}
				}
			}
			// execute the method
			return next(ctx, request)
		}

		return fn
	}

	unaryFn := connect.UnaryInterceptorFunc(interFn)
	// prepare the option
	return connect.WithInterceptors(unaryFn)
}
