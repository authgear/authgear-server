package flows

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideUserController(
	u UserProvider,
	ti TokenIssuer,
	scc session.CookieConfiguration,
	sp session.Provider,
	hp hook.Provider,
	tp time.Provider,
	c *config.TenantConfiguration,
) *UserController {
	return &UserController{
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

func ProvideWebAppFlow(
	c *config.TenantConfiguration,
	idp IdentityProvider,
	up UserProvider,
	hp hook.Provider,
	ip InteractionProvider,
	uc *UserController,
) *WebAppFlow {
	return &WebAppFlow{
		ConflictConfig: c.AppConfig.Identity.OnConflict,
		Identities:     idp,
		Users:          up,
		Hooks:          hp,
		Interactions:   ip,
		UserController: uc,
	}
}

var DependencySet = wire.NewSet(
	ProvideWebAppFlow,
	wire.Struct(new(AnonymousFlow), "*"),
	wire.Struct(new(PasswordFlow), "*"),
	ProvideIsAnonymousIdentityEnabled,
	ProvideUserController,
)
