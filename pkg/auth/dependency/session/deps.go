package session

import (
	"net/http"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type InsecureCookieConfig bool

func ProvideSessionCookieConfiguration(
	r *http.Request,
	icc InsecureCookieConfig,
	c *config.TenantConfiguration,
) CookieConfiguration {
	return NewSessionCookieConfiguration(r, bool(icc), *c.AppConfig.Session)
}

func ProvideSessionMiddleware(cfg CookieConfiguration, r Resolver, ais authinfo.Store, tx db.TxContext) *Middleware {
	return &Middleware{
		CookieConfiguration: cfg,
		SessionResolver:     r,
		AuthInfoStore:       ais,
		TxContext:           tx,
	}
}

func ProvideSessionProvider(req *http.Request, s Store, es EventStore, c *config.TenantConfiguration) Provider {
	return NewProvider(req, s, es, *c.AppConfig.Session)
}

var DependencySet = wire.NewSet(
	ProvideSessionCookieConfiguration,
	wire.Bind(new(Resolver), new(Provider)),
	ProvideSessionMiddleware,
	ProvideSessionProvider,
)
