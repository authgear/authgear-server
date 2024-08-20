package webapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SettingImplementationSwitcherMiddleware struct {
	UIConfig *config.UIConfig
}

type settingImplementationSwitcherContextKeyType struct{}

var settingImplementationSwitcherContextKey = settingImplementationSwitcherContextKeyType{}

type settingImplementationSwitcherContext struct {
	SettingUIImplementation config.SettingUIImplementation
}

func WithSettingUIImplementation(ctx context.Context, impl config.SettingUIImplementation) context.Context {
	v, ok := ctx.Value(settingImplementationSwitcherContextKey).(*settingImplementationSwitcherContext)
	if ok {
		v.SettingUIImplementation = impl
		return ctx
	}

	return context.WithValue(ctx, settingImplementationSwitcherContextKey, &settingImplementationSwitcherContext{
		SettingUIImplementation: impl,
	})
}

func GetSettingUIImplementation(ctx context.Context) config.SettingUIImplementation {
	v, ok := ctx.Value(settingImplementationSwitcherContextKey).(*settingImplementationSwitcherContext)
	if !ok {
		return config.SettingUIImplementationDefault
	}
	return v.SettingUIImplementation
}

func (m *SettingImplementationSwitcherMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(WithSettingUIImplementation(r.Context(), m.UIConfig.SettingImplementation))
		next.ServeHTTP(w, r)
	})
}

type SettingImplementationSwitcherHandler struct {
	SettingV1 http.Handler
	SettingV2 http.Handler
}

func (h *SettingImplementationSwitcherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	impl := GetSettingUIImplementation(r.Context()).WithDefault()
	switch impl {
	case config.SettingUIImplementationV1:
		h.SettingV1.ServeHTTP(w, r)
	case config.SettingUIImplementationV2:
		h.SettingV2.ServeHTTP(w, r)
	default:
		panic(fmt.Errorf("unknown setting ui implementation %s", impl))
	}
}
