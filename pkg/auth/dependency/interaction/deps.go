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
	"github.com/skygeario/skygear-server/pkg/core/template"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideOOBProvider(
	c *config.TenantConfiguration,
	te *template.Engine,
	upp urlprefix.Provider,
	tq async.Queue,
) *OOBProviderImpl {
	return &OOBProviderImpl{
		SMSMessageConfiguration:       c.AppConfig.Messages.SMS,
		EmailMessageConfiguration:     c.AppConfig.Messages.Email,
		AuthenticatorOOBConfiguration: c.AppConfig.Authenticator.OOB,
		TemplateEngine:                te,
		URLPrefixProvider:             upp,
		TaskQueue:                     tq,
	}
}

func ProvideProvider(
	s Store,
	t time.Provider,
	lf logging.Factory,
	ip IdentityProvider,
	ap AuthenticatorProvider,
	up UserProvider,
	oob OOBProvider,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Store:         s,
		Time:          t,
		Logger:        lf.NewLogger("interaction"),
		Identity:      ip,
		Authenticator: ap,
		User:          up,
		OOB:           oob,
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
) UserProvider {
	return &userProvider{
		AuthInfos:                     ais,
		UserProfiles:                  ups,
		Time:                          tp,
		Hooks:                         hp,
		URLPrefix:                     up,
		TaskQueue:                     q,
		WelcomeEmailConfiguration:     config.AppConfig.WelcomeEmail,
		UserVerificationConfiguration: config.AppConfig.UserVerification,
	}
}

var DependencySet = wire.NewSet(
	ProvideOOBProvider,
	wire.Bind(new(OOBProvider), new(*OOBProviderImpl)),
	ProvideProvider,
	ProvideUserProvider,
)
