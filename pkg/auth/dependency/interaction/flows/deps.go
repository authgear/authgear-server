package flows

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideUserController(
	ais authinfo.Store,
	u UserProvider,
	ti TokenIssuer,
	scc session.CookieConfiguration,
	sp session.Provider,
	hp hook.Provider,
	tp time.Provider,
	c *config.TenantConfiguration,
) *UserController {
	return &UserController{
		AuthInfos:           ais,
		Users:               u,
		TokenIssuer:         ti,
		SessionCookieConfig: scc,
		Sessions:            sp,
		Hooks:               hp,
		Time:                tp,
		Clients:             c.AppConfig.Clients,
	}
}

type IsAnonymousIdentityEnabled bool

func ProvideIsAnonymousIdentityEnabled(c *config.TenantConfiguration) IsAnonymousIdentityEnabled {
	for _, i := range c.AppConfig.Authentication.Identities {
		if i == string(authn.IdentityTypeAnonymous) {
			return true
		}
	}
	return false
}

var DependencySet = wire.NewSet(
	wire.Struct(new(WebAppFlow), "*"),
	wire.Struct(new(AnonymousFlow), "*"),
	wire.Struct(new(PasswordFlow), "*"),
	ProvideIsAnonymousIdentityEnabled,
	ProvideUserController,
)
