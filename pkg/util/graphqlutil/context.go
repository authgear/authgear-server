package graphqlutil

import (
	"context"
)

type GraphQLContext interface{}

type contextKeyType struct{}

var contextKey = contextKeyType{}

func WithContext(ctx context.Context, gqlContext GraphQLContext) context.Context {
	return context.WithValue(ctx, contextKey, gqlContext)
}

func GQLContext(ctx context.Context) GraphQLContext {
	return ctx.Value(contextKey).(GraphQLContext)
}
