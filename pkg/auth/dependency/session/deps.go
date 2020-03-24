package session

import (
	"net/http"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type InsecureCookieConfig bool

func ProvideSessionCookieConfiguration(
	r *http.Request,
	icc InsecureCookieConfig,
	c *config.TenantConfiguration,
) CookieConfiguration {
	return NewSessionCookieConfiguration(r, bool(icc), *c.AppConfig.Session)
}

func ProvideSessionResolver(cfg CookieConfiguration, p ResolverProvider) *Resolver {
	return &Resolver{
		CookieConfiguration: cfg,
		Provider:            p,
	}
}

func ProvideSessionProvider(req *http.Request, s Store, aep auth.AccessEventProvider, c *config.TenantConfiguration) Provider {
	return NewProvider(req, s, aep, *c.AppConfig.Session)
}

var DependencySet = wire.NewSet(
	ProvideSessionCookieConfiguration,
	ProvideSessionProvider,
	wire.Bind(new(ResolverProvider), new(Provider)),
	ProvideSessionResolver,
	wire.Bind(new(auth.IDPSessionResolver), new(*Resolver)),
)
