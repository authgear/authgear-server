package webapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SettingsImplementationSwitcherMiddlewareUIImplementationService interface {
	GetSettingsUIImplementation() config.SettingsUIImplementation
}

type SettingsImplementationSwitcherMiddleware struct {
	UIImplementationService SettingsImplementationSwitcherMiddlewareUIImplementationService
}

type settingsImplementationSwitcherContextKeyType struct{}

var settingsImplementationSwitcherContextKey = settingsImplementationSwitcherContextKeyType{}

type settingsImplementationSwitcherContext struct {
	SettingsUIImplementation config.SettingsUIImplementation
}

func WithSettingsUIImplementation(ctx context.Context, impl config.SettingsUIImplementation) context.Context {
	v, ok := ctx.Value(settingsImplementationSwitcherContextKey).(*settingsImplementationSwitcherContext)
	if ok {
		v.SettingsUIImplementation = impl
		return ctx
	}

	return context.WithValue(ctx, settingsImplementationSwitcherContextKey, &settingsImplementationSwitcherContext{
		SettingsUIImplementation: impl,
	})
}

func GetSettingsUIImplementation(ctx context.Context) config.SettingsUIImplementation {
	v, ok := ctx.Value(settingsImplementationSwitcherContextKey).(*settingsImplementationSwitcherContext)
	if !ok {
		return config.SettingsUIImplementationDefault
	}
	return v.SettingsUIImplementation
}

func (m *SettingsImplementationSwitcherMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := m.UIImplementationService.GetSettingsUIImplementation()
		r = r.WithContext(WithSettingsUIImplementation(r.Context(), val))
		next.ServeHTTP(w, r)
	})
}

type SettingsImplementationSwitcherHandler struct {
	SettingV1 http.Handler
	SettingV2 http.Handler
}

func (h *SettingsImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	impl := GetSettingsUIImplementation(r.Context())
	switch impl {
	case config.SettingsUIImplementationV1:
		h.SettingV1.ServeHTTP(w, r)
	case config.SettingsUIImplementationV2:
		h.SettingV2.ServeHTTP(w, r)
	default:
		panic(fmt.Errorf("unknown setting ui implementation %s", impl))
	}
}
