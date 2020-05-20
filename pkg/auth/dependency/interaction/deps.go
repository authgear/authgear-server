package interaction

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	s Store,
	t time.Provider,
	lf logging.Factory,
	ip IdentityProvider,
	ap AuthenticatorProvider,
	up UserProvider,
	oob OOBProvider,
	c *config.TenantConfiguration,
	hp hook.Provider,
) *Provider {
	return &Provider{
		Store:         s,
		Time:          t,
		Logger:        lf.NewLogger("interaction"),
		Identity:      ip,
		Authenticator: ap,
		User:          up,
		OOB:           oob,
		Hooks:         hp,
		Config:        c.AppConfig.Authentication,
	}
}

func ProvideUserProvider(
	ais authinfo.Store,
	ups userprofile.Store,
	tp time.Provider,
	hp hook.Provider,
	up urlprefix.Provider,
	q async.Queue,
	config *config.TenantConfiguration,
	wmp WelcomeMessageProvider,
) UserProvider {
	return &userProvider{
		AuthInfos:                     ais,
		UserProfiles:                  ups,
		Time:                          tp,
		Hooks:                         hp,
		URLPrefix:                     up,
		TaskQueue:                     q,
		UserVerificationConfiguration: config.AppConfig.UserVerification,
		WelcomeMessageProvider:        wmp,
	}
}

var DependencySet = wire.NewSet(
	ProvideProvider,
	ProvideUserProvider,
)
