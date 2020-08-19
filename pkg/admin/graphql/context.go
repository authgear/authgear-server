package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
)

type UserLoader interface {
	Get(id string) (*user.User, error)
	QueryPage(args loader.PageArgs) (*loader.PageResult, error)
}

type Context struct {
	Users UserLoader
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return context.WithValue(ctx, contextKey, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return ctx.Value(contextKey).(*Context)
}
