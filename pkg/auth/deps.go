package auth

import (
	"context"
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	authredis "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	authenticatorbearertoken "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/bearertoken"
	authenticatoroob "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	authenticatorpassword "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	authenticatorprovider "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/provider"
	authenticatorrecoverycode "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/recoverycode"
	authenticatortotp "github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/challenge"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpassword"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	identityanonymous "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	identityloginid "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	identityoauth "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	identityprovider "github.com/skygeario/skygear-server/pkg/auth/dependency/identity/provider"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	interactionredis "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	mfapq "github.com/skygeario/skygear-server/pkg/auth/dependency/mfa/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	oauthhandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	oauthpq "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/pq"
	oauthredis "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	oidchandler "github.com/skygeario/skygear-server/pkg/auth/dependency/oidc/handler"
	passwordhistorypq "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	oauthprincipal "github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	passwordprincipal "github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	sessionredis "github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"
	"github.com/skygeario/skygear-server/pkg/auth/deps"
	"github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	authinfopq "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	coretemplate "github.com/skygeario/skygear-server/pkg/core/template"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func MakeHandler(deps DependencyMap, factory func(r *http.Request, m DependencyMap) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := factory(r, deps)
		h.ServeHTTP(w, r)
	})
}

func MakeMiddleware(deps DependencyMap, factory func(r *http.Request, m DependencyMap) mux.MiddlewareFunc) mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := factory(r, deps)
			h := m(next)
			h.ServeHTTP(w, r)
		})
	})
}

func ProvideContext(r *http.Request) context.Context {
	return r.Context()
}

func ProvideLoggingRequestID(r *http.Request) logging.RequestID {
	return logging.RequestID(r.Header.Get(corehttp.HeaderRequestID))
}

func ProvideTenantConfig(ctx context.Context, m DependencyMap) *config.TenantConfiguration {
	// populate default
	tc := config.GetTenantConfig(ctx)
	appConfig := *tc.AppConfig
	if !appConfig.SMTP.IsValid() {
		appConfig.SMTP = m.DefaultConfiguration.SMTP
	}

	if !appConfig.Twilio.IsValid() {
		appConfig.Twilio = m.DefaultConfiguration.Twilio
	}

	if !appConfig.Nexmo.IsValid() {
		appConfig.Nexmo = m.DefaultConfiguration.Nexmo
	}
	tConfig := *tc
	tConfig.AppConfig = &appConfig
	return &tConfig
}

func ProvideSessionInsecureCookieConfig(m DependencyMap) session.InsecureCookieConfig {
	return session.InsecureCookieConfig(m.UseInsecureCookie)
}

func ProvideMFAInsecureCookieConfig(m DependencyMap) mfa.InsecureCookieConfig {
	return mfa.InsecureCookieConfig(m.UseInsecureCookie)
}

func ProvideValidator(m DependencyMap) *validation.Validator {
	return m.Validator
}

func ProvideReservedNameChecker(m DependencyMap) *loginid.ReservedNameChecker {
	return m.ReservedNameChecker
}

func ProvideTaskExecutor(m DependencyMap) *async.Executor {
	return m.AsyncTaskExecutor
}

func ProvideTemplateEngine(config *config.TenantConfiguration, m DependencyMap) *coretemplate.Engine {
	return template.NewEngineWithConfig(
		*config,
		m.EnableFileSystemTemplate,
		m.AssetGearLoader,
	)
}

func ProvideAuthSQLBuilder(f db.SQLBuilderFactory) db.SQLBuilder {
	return f("auth")
}

func ProvidePrincipalProviders(
	oauth oauthprincipal.Provider,
	password passwordprincipal.Provider,
) []principal.Provider {
	return []principal.Provider{oauth, password}
}

func ProvideStaticAssetURLPrefix(m DependencyMap) deps.StaticAssetURLPrefix {
	return deps.StaticAssetURLPrefix(m.StaticAssetURLPrefix)
}

func ProvideCSRFMiddleware(m DependencyMap, tConfig *config.TenantConfiguration) mux.MiddlewareFunc {
	middleware := &webapp.CSRFMiddleware{
		// NOTE(webapp): reuse Authentication.Secret instead of creating a new one.
		Key:               tConfig.AppConfig.Authentication.Secret,
		UseInsecureCookie: m.UseInsecureCookie,
	}
	return middleware.Handle
}

var interactionDependencySet = wire.NewSet(
	identityloginid.DependencySet,
	identityoauth.DependencySet,
	identityanonymous.DependencySet,
	identityprovider.DependencySet,
	wire.Bind(new(webapp.IdentityProvider), new(*identityprovider.Provider)),
	authenticatorpassword.DependencySet,
	authenticatortotp.DependencySet,
	authenticatoroob.DependencySet,
	authenticatorbearertoken.DependencySet,
	authenticatorrecoverycode.DependencySet,
	authenticatorprovider.DependencySet,
	interaction.DependencySet,
	interactionredis.DependencySet,
	interactionflows.DependencySet,

	wire.Bind(new(interaction.OOBProvider), new(*authenticatoroob.Provider)),
	wire.Bind(new(interaction.IdentityProvider), new(*identityprovider.Provider)),
	wire.Bind(new(interaction.AuthenticatorProvider), new(*authenticatorprovider.Provider)),
	wire.Bind(new(identityprovider.LoginIDIdentityProvider), new(*identityloginid.Provider)),
	wire.Bind(new(identityprovider.OAuthIdentityProvider), new(*identityoauth.Provider)),
	wire.Bind(new(identityprovider.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
	wire.Bind(new(authenticatorprovider.PasswordAuthenticatorProvider), new(*authenticatorpassword.Provider)),
	wire.Bind(new(authenticatorprovider.TOTPAuthenticatorProvider), new(*authenticatortotp.Provider)),
	wire.Bind(new(authenticatorprovider.OOBOTPAuthenticatorProvider), new(*authenticatoroob.Provider)),
	wire.Bind(new(authenticatorprovider.BearerTokenAuthenticatorProvider), new(*authenticatorbearertoken.Provider)),
	wire.Bind(new(authenticatorprovider.RecoveryCodeAuthenticatorProvider), new(*authenticatorrecoverycode.Provider)),

	wire.Bind(new(interactionflows.InteractionProvider), new(*interaction.Provider)),
	wire.Bind(new(interactionflows.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

	wire.Bind(new(webapp.InteractionFlow), new(*interactionflows.WebAppFlow)),
	wire.Bind(new(oauthhandler.AnonymousInteractionFlow), new(*interactionflows.AnonymousFlow)),
	wire.Bind(new(webapp.AnonymousFlow), new(*interactionflows.AnonymousFlow)),

	wire.Bind(new(hook.LoginIDProvider), new(*identityloginid.Provider)),

	wire.Bind(new(forgotpassword.LoginIDProvider), new(*identityloginid.Provider)),
	wire.Bind(new(forgotpassword.ResetPasswordFlow), new(*interactionflows.PasswordFlow)),
)

var challengeDependencySet = wire.NewSet(
	challenge.DependencySet,
	wire.Bind(new(interactionflows.ChallengeProvider), new(*challenge.Provider)),
)

var endpointsDependencySet = wire.NewSet(
	endpointsProviderSet,

	wire.Bind(new(oauth.AuthorizeEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oauth.TokenEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oauth.RevokeEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oidc.JWKSEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oidc.UserInfoEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oidc.EndSessionEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oauthhandler.EndpointsProvider), new(*EndpointsProvider)),
	wire.Bind(new(webapp.EndpointsProvider), new(*EndpointsProvider)),
)

var webappDependencySet = wire.NewSet(
	webapp.DependencySet,

	wire.Bind(new(oauthhandler.AuthenticateURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(oidchandler.LogoutURLProvider), new(*webapp.URLProvider)),
	wire.Bind(new(oidchandler.SettingsURLProvider), new(*webapp.URLProvider)),
)

var CommonDependencySet = wire.NewSet(
	ProvideTenantConfig,
	ProvideSessionInsecureCookieConfig,
	ProvideMFAInsecureCookieConfig,
	ProvideValidator,
	ProvideReservedNameChecker,
	ProvideTaskExecutor,
	ProvideTemplateEngine,
	ProvideStaticAssetURLPrefix,
	endpointsDependencySet,

	ProvideAuthSQLBuilder,
	ProvidePrincipalProviders,

	logging.DependencySet,
	time.DependencySet,
	db.DependencySet,
	authinfopq.DependencySet,
	userprofile.DependencySet,
	session.DependencySet,
	sessionredis.DependencySet,
	handler.DependencySet,
	coreauth.DependencySet,
	async.DependencySet,
	// TODO(deps): Remove sms and mail from CommonDependencySet
	// to prevent their use in HTTP request.
	sms.DependencySet,
	mail.DependencySet,

	hook.DependencySet,
	auth.DependencySet,
	authredis.DependencySet,
	authn.DependencySet,
	audit.DependencySet,
	loginid.DependencySet,
	passwordhistorypq.DependencySet,
	principal.DependencySet,
	oauthprincipal.DependencySet,
	passwordprincipal.DependencySet,
	sso.DependencySet,
	urlprefix.DependencySet,
	mfa.DependencySet,
	mfapq.DependencySet,
	webappDependencySet,
	oauthhandler.DependencySet,
	oauth.DependencySet,
	oauthpq.DependencySet,
	oauthredis.DependencySet,
	oidc.DependencySet,
	oidchandler.DependencySet,
	welcemail.DependencySet,
	userverify.DependencySet,
	forgotpassword.DependencySet,
	challengeDependencySet,
	interactionDependencySet,
)

// DependencySet is for HTTP request
var DependencySet = wire.NewSet(
	CommonDependencySet,
	ProvideContext,
	ProvideLoggingRequestID,
)
