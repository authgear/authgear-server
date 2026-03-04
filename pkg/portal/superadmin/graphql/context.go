package graphql

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Context struct {
	Request *http.Request
	// Add service interfaces here as operations are added
}

func WithContext(ctx context.Context, gqlCtx *Context) context.Context {
	return graphqlutil.WithContext(ctx, gqlCtx)
}

func GQLContext(ctx context.Context) *Context {
	return graphqlutil.GQLContext(ctx).(*Context)
}
