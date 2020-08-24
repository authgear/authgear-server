package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type ViewerLoader interface {
	Get() *graphqlutil.Lazy
}

type AppLoader interface {
	Get(id string) *graphqlutil.Lazy
	QueryPage(args graphqlutil.PageArgs) (*graphqlutil.PageResult, error)
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

type Context struct {
	Viewer ViewerLoader
	Apps   AppLoader
}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return context.WithValue(ctx, contextKey, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return ctx.Value(contextKey).(*Context)
}
