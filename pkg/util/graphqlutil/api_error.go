package graphqlutil

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type APIErrorExtension struct {
}

func (A APIErrorExtension) Name() string {
	return "APIError"
}

func (A APIErrorExtension) Init(ctx context.Context, params *graphql.Params) context.Context {
	return ctx
}

func (A APIErrorExtension) ParseDidStart(ctx context.Context) (context.Context, graphql.ParseFinishFunc) {
	return ctx, func(err error) {}
}

func (A APIErrorExtension) ValidationDidStart(ctx context.Context) (context.Context, graphql.ValidationFinishFunc) {
	return ctx, func(errors []gqlerrors.FormattedError) {}
}

func (A APIErrorExtension) ExecutionDidStart(ctx context.Context) (context.Context, graphql.ExecutionFinishFunc) {
	return ctx, func(result *graphql.Result) {
		for i, gqlError := range result.Errors {
			apiError := apierrors.AsAPIError(originalError(gqlError))
			if apiError != nil {
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
}

func (A APIErrorExtension) ResolveFieldDidStart(ctx context.Context, info *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc) {
	return ctx, func(i interface{}, err error) {}
}

func (A APIErrorExtension) HasResult() bool { return false }

func (A APIErrorExtension) GetResult(ctx context.Context) interface{} { return nil }

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
