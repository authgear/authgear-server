package graphql

import (
	"context"
)

type ViewerLoader interface {
	Get() (interface{}, error)
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

type Context struct {
	Viewer ViewerLoader
}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return context.WithValue(ctx, contextKey, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return ctx.Value(contextKey).(*Context)
}
