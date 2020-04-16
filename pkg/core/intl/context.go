package intl

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type intlContext struct {
	PreferredLanguageTags []string
}

func WithPreferredLanguageTags(ctx context.Context, tags []string) context.Context {
	v, ok := ctx.Value(contextKey).(*intlContext)
	if ok {
		v.PreferredLanguageTags = tags
		return ctx
	}

	return context.WithValue(ctx, contextKey, &intlContext{
		PreferredLanguageTags: tags,
	})
}

func GetPreferredLanguageTags(ctx context.Context) []string {
	v, ok := ctx.Value(contextKey).(*intlContext)
	if !ok {
		return nil
	}
	return v.PreferredLanguageTags
}
