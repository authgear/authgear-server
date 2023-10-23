package webapp

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type ImplementationSwitcherMiddleware struct {
	UIConfig *config.UIConfig
}

type implementationSwitcherContextKeyType struct{}

var implementationSwitcherContextKey = implementationSwitcherContextKeyType{}

type implementationSwitcherContext struct {
	UIImplementation config.UIImplementation
}

func WithUIImplementation(ctx context.Context, impl config.UIImplementation) context.Context {
	v, ok := ctx.Value(implementationSwitcherContextKey).(*implementationSwitcherContext)
	if ok {
		v.UIImplementation = impl
		return ctx
	}

	return context.WithValue(ctx, implementationSwitcherContextKey, &implementationSwitcherContext{
		UIImplementation: impl,
	})
}

func GetUIImplementation(ctx context.Context) config.UIImplementation {
	v, ok := ctx.Value(implementationSwitcherContextKey).(*implementationSwitcherContext)
	if !ok {
		return ""
	}
	return v.UIImplementation
}

func (m *ImplementationSwitcherMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(WithUIImplementation(r.Context(), m.UIConfig.Implementation))
		next.ServeHTTP(w, r)
	})
}

type ImplementationSwitcherHandler struct {
	Interaction http.Handler
	Authflow    http.Handler
}

func (h *ImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch GetUIImplementation(r.Context()) {
	case config.UIImplementationAuthflow:
		h.Authflow.ServeHTTP(w, r)
	default:
		h.Interaction.ServeHTTP(w, r)
	}
}
