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

// Singleton returns the named dependency from root context, and construct it
// if not yet created.
// This function should not be used at the moment, since we don't have a root
// context yet, so effectively it is same as Scoped.
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

// Transient returns a newly created named dependency, existing singleton and
// scoped dependencies are ignored.
func Transient(ctx context.Context, name string, factory func() interface{}) func() interface{} {
	return func() interface{} { return factory() }
}

// Scoped returns the named dependency from current context, and construct
// it if not yet created.
// This function should be used in DependencyMap.Provide function only, since
// we do not update the request context yet.
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
