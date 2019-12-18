package inject

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type injectContext struct {
	Root         *injectContext
	Parent       *injectContext
	Dependencies map[string]interface{}
}

func WithInject(ctx context.Context) context.Context {
	injectCtx := &injectContext{Dependencies: map[string]interface{}{}}

	if baseCtx, ok := ctx.Value(contextKey).(*injectContext); ok {
		injectCtx.Parent = baseCtx
		injectCtx.Root = baseCtx.Root
	} else {
		injectCtx.Parent = nil
		injectCtx.Root = injectCtx
	}

	return context.WithValue(ctx, contextKey, injectCtx)
}

func getContext(ctx context.Context) *injectContext {
	injectCtx, _ := ctx.Value(contextKey).(*injectContext)
	return injectCtx
}

func Singleton(ctx context.Context, name string, factory func() interface{}) func() interface{} {
	return func() interface{} {
		injectCtx := getContext(ctx)
		if injectCtx != nil {
			if dep, ok := injectCtx.Root.Dependencies[name]; ok {
				return dep
			}
		}

		dep := factory()
		if injectCtx != nil {
			injectCtx.Root.Dependencies[name] = dep
		}
		return dep
	}
}

func Transient(ctx context.Context, name string, factory func() interface{}) func() interface{} {
	return func() interface{} { return factory() }
}

func Scoped(ctx context.Context, name string, factory func() interface{}) func() interface{} {
	return func() interface{} {
		injectCtx := getContext(ctx)
		if injectCtx != nil {
			if dep, ok := injectCtx.Dependencies[name]; ok {
				return dep
			}
		}

		dep := factory()
		if injectCtx != nil {
			injectCtx.Dependencies[name] = dep
		}
		return dep
	}
}
