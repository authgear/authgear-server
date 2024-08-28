package webapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SettingsImplementationSwitcherMiddleware struct {
	UIConfig *config.UIConfig
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
		r = r.WithContext(WithSettingsUIImplementation(r.Context(), m.UIConfig.SettingsImplementation))
		next.ServeHTTP(w, r)
	})
}

type SettingsImplementationSwitcherHandler struct {
	SettingV1 http.Handler
	SettingV2 http.Handler
}

func (h *SettingsImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	impl := GetSettingsUIImplementation(r.Context()).WithDefault()
	switch impl {
	case config.SettingsUIImplementationV1:
		h.SettingV1.ServeHTTP(w, r)
	case config.SettingsUIImplementationV2:
		h.SettingV2.ServeHTTP(w, r)
	default:
		panic(fmt.Errorf("unknown setting ui implementation %s", impl))
	}
}
