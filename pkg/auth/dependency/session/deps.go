package session

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type InsecureCookieConfig bool

func ProvideSessionCookieConfiguration(
	r *http.Request,
	icc InsecureCookieConfig,
	c *config.TenantConfiguration,
) CookieDef {
	return NewSessionCookieDef(r, bool(icc), *c.AppConfig.Session)
}

func ProvideSessionProvider(req *http.Request, s Store, aep *auth.AccessEventProvider, c *config.TenantConfiguration) Provider {
	return NewProvider(req, s, aep, *c.AppConfig.Session)
}

func ProvideSessionManager(s Store, tp time.Provider, c *config.TenantConfiguration, cc CookieDef) *Manager {
	return &Manager{
		Store:     s,
		Time:      tp,
		Config:    *c.AppConfig.Session,
		CookieDef: cc,
	}
}

var DependencySet = wire.NewSet(
	ProvideSessionCookieConfiguration,
	ProvideSessionProvider,
	wire.Bind(new(ResolverProvider), new(Provider)),
	wire.Struct(new(Resolver), "*"),
	wire.Bind(new(auth.IDPSessionResolver), new(*Resolver)),
	ProvideSessionManager,
	wire.Bind(new(auth.IDPSessionManager), new(*Manager)),
)
