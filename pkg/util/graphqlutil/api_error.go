package graphqlutil

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type APIErrorExtension struct{}

func (a APIErrorExtension) Name() string {
	return "APIError"
}

func (a APIErrorExtension) Init(ctx context.Context, params *graphql.Params) context.Context {
	return ctx
}

func (a APIErrorExtension) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	return ctx, func(err error) {}
}

func (a APIErrorExtension) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	return ctx, func(errors []gqlerrors.FormattedError) {}
}

func (a APIErrorExtension) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	return ctx, func(result *graphql.Result) {
		logger := GQLContext(ctx).Logger()
		for i, gqlError := range result.Errors {
			err := originalError(gqlError)

			// This error will appear when GraphiQL is opened, it is better to ignore this.
			// I know it is not a good practice to match on the error message,
			// but the original error is constructed with fmt.Errorf inline.
			// So I have no choice :(
			// See https://github.com/graphql-go/graphql/blob/1a9db8859ef57c2821bbd47b0db9a1a09e617f41/executor.go#L144
			if err.Error() == "Must provide an operation." {
				continue
			}

			if !apierrors.IsAPIError(err) {
				// FIXME(graphql): Log panics correctly
				//		graphql-go recovers panic and translates it to error automatically.
				//		However, if the panic value is not of type `error` or `string`,
				// 		it just use a generic error message.
				//     	For some panics, string concatenation leads to the compiler
				//		infers a string enum type, and therefore cannot be logged here.
				logger.
					WithError(err).
					WithField("path", gqlError.Path).
					Error("unexpected error when executing GraphQL query")
				continue
			}

			apiError := apierrors.AsAPIError(err)
			if gqlError.Extensions == nil {
				gqlError.Extensions = make(map[string]interface{})
			}
			gqlError.Extensions["errorName"] = apiError.Name
			gqlError.Extensions["reason"] = apiError.Reason
			if len(apiError.Info) > 0 {
				gqlError.Extensions["info"] = apiError.Info
			}
			result.Errors[i] = gqlError
		}
	}
}

func (a APIErrorExtension) ResolveFieldDidStart(ctx context.Context, info *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	return ctx, func(i interface{}, err error) {}
}

func (a APIErrorExtension) HasResult() bool { return false }

func (a APIErrorExtension) GetResult(ctx context.Context) interface{} { return nil }

func originalError(err error) error {
	for err != nil {
		if wrapper, ok := err.(interface{ OriginalError() error }); ok {
			err = wrapper.OriginalError()
		} else if gError, ok := err.(*gqlerrors.Error); ok {
			err = gError.OriginalError
		} else {
			break
		}
	}
	return err
}
