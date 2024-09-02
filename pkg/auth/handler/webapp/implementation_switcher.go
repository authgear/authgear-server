package webapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type ImplementationSwitcherMiddlewareUIImplementationService interface {
	GetUIImplementation() config.UIImplementation
}

type ImplementationSwitcherMiddleware struct {
	UIImplementationService ImplementationSwitcherMiddlewareUIImplementationService
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
		return config.UIImplementationDefault
	}
	return v.UIImplementation
}

func (m *ImplementationSwitcherMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uiImpl := m.UIImplementationService.GetUIImplementation()
		r = r.WithContext(WithUIImplementation(r.Context(), uiImpl))
		next.ServeHTTP(w, r)
	})
}

type ImplementationSwitcherHandler struct {
	Interaction http.Handler
	Authflow    http.Handler
	AuthflowV2  http.Handler
}

func (h *ImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	impl := GetUIImplementation(r.Context())
	switch impl {
	case config.UIImplementationAuthflow:
		h.Authflow.ServeHTTP(w, r)
	case config.UIImplementationAuthflowV2:
		h.AuthflowV2.ServeHTTP(w, r)
	case config.UIImplementationInteraction:
		h.Interaction.ServeHTTP(w, r)
	default:
		panic(fmt.Errorf("unknown ui implementation %s", impl))
	}
}
