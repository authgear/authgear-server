package webapp

import (
	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(ValidateProviderImpl), "*"),
	wire.Struct(new(RenderProviderImpl), "*"),
	wire.Struct(new(StateStoreImpl), "*"),
	wire.Bind(new(StateStore), new(*StateStoreImpl)),
	wire.Struct(new(StateProviderImpl), "*"),
	wire.Bind(new(StateProvider), new(*StateProviderImpl)),
	wire.Struct(new(URLProvider), "*"),
)

func ProvideCSPMiddleware(c *config.OAuthConfig) mux.MiddlewareFunc {
	m := &CSPMiddleware{Clients: c.Clients}
	return m.Handle
}

func ProvideStateMiddleware(stateStore StateStore) mux.MiddlewareFunc {
	m := &StateMiddleware{StateStore: stateStore}
	return m.Handle
}
