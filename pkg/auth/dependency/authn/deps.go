package authn

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideSignupProcess(
	pc *audit.PasswordChecker,
	lc loginid.LoginIDChecker,
	ip principal.IdentityProvider,
	pp password.Provider,
	op oauth.Provider,
	tp coreTime.Provider,
	as authinfo.Store,
	us userprofile.Store,
	hp hook.Provider,
	cfg *config.TenantConfiguration,
	up urlprefix.Provider,
	q async.Queue,
) *SignupProcess {
	return &SignupProcess{
		PasswordChecker:               pc,
		LoginIDChecker:                lc,
		IdentityProvider:              ip,
		PasswordProvider:              pp,
		OAuthProvider:                 op,
		TimeProvider:                  tp,
		AuthInfoStore:                 as,
		UserProfileStore:              us,
		HookProvider:                  hp,
		WelcomeEmailConfiguration:     cfg.AppConfig.WelcomeEmail,
		UserVerificationConfiguration: cfg.AppConfig.UserVerification,
		LoginIDConflictConfiguration:  cfg.AppConfig.AuthAPI.OnIdentityConflict.LoginID,
		URLPrefixProvider:             up,
		TaskQueue:                     q,
	}
}

func ProvideAuthenticateProcess(
	loggerFactory logging.Factory,
	tp coreTime.Provider,
	pp password.Provider,
	op oauth.Provider,
	ip principal.IdentityProvider,
) *AuthenticateProcess {
	return &AuthenticateProcess{
		Logger:           loggerFactory.NewLogger("authn-process"),
		TimeProvider:     tp,
		PasswordProvider: pp,
		OAuthProvider:    op,
		IdentityProvider: ip,
	}
}

func ProvideSessionProvider(
	mp mfa.Provider,
	sp session.Provider,
	cfg *config.TenantConfiguration,
	tp time.Provider,
	as authinfo.Store,
	us userprofile.Store,
	ip principal.IdentityProvider,
	hp hook.Provider,
	ti TokenIssuer,
) *SessionProvider {
	return &SessionProvider{
		MFAProvider:        mp,
		SessionProvider:    sp,
		ClientConfigs:      cfg.AppConfig.Clients,
		MFAConfig:          cfg.AppConfig.MFA,
		AuthnSessionConfig: cfg.AppConfig.Auth.AuthenticationSession,
		TimeProvider:       tp,
		AuthInfoStore:      as,
		UserProfileStore:   us,
		IdentityProvider:   ip,
		HookProvider:       hp,
		TokenIssuer:        ti,
	}
}

var DependencySet = wire.NewSet(
	ProvideSignupProcess,
	ProvideAuthenticateProcess,
	wire.Struct(new(OAuthCoordinator), "*"),
	ProvideSessionProvider,
	wire.Struct(new(ProviderFactory), "*"),
)

func ProvideAuthAPIProvider(f *ProviderFactory) *Provider { return f.ForAuthAPI() }
func ProvideAuthUIProvider(f *ProviderFactory) *Provider  { return f.ForAuthUI() }
