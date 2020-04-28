//+build wireinject

package sso

import (
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"
	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func provideOAuthProviderFromRequestVars(r *http.Request, spf *sso.OAuthProviderFactory) sso.OAuthProvider {
	vars := mux.Vars(r)
	return spf.NewOAuthProvider(vars["provider"])
}

func ProvideRedirectURIForAPIFunc() sso.RedirectURLFunc {
	return RedirectURIForAPI
}

func provideAuthHandler(
	tx db.TxContext,
	cfg *config.TenantConfiguration,
	hp sso.AuthHandlerHTMLProvider,
	sp sso.Provider,
	ap AuthHandlerAuthnProvider,
	op sso.OAuthProvider,
	f OAuthHandlerInteractionFlow,
) http.Handler {
	h := &AuthHandler{
		TxContext:               tx,
		TenantConfiguration:     cfg,
		AuthHandlerHTMLProvider: hp,
		SSOProvider:             sp,
		AuthnProvider:           ap,
		OAuthProvider:           op,
		Interactions:            f,
	}
	return h
}

func newAuthHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authn.ProvideAuthAPIProvider,
		wire.Bind(new(AuthHandlerAuthnProvider), new(*authn.Provider)),
		provideOAuthProviderFromRequestVars,
		provideAuthHandler,
		ProvideRedirectURIForAPIFunc,
		wire.Bind(new(OAuthHandlerInteractionFlow), new(*interactionflows.AuthAPIFlow)),
	)
	return nil
}

func provideAuthResultHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	ap AuthResultAuthnProvider,
	v *validation.Validator,
	sp sso.Provider,
	f OAuthResultInteractionFlow,
) http.Handler {
	h := &AuthResultHandler{
		TxContext:     tx,
		AuthnProvider: ap,
		Validator:     v,
		SSOProvider:   sp,
		Interactions:  f,
	}
	return requireAuthz(h, h)
}

func newAuthResultHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authn.ProvideAuthAPIProvider,
		wire.Bind(new(AuthResultAuthnProvider), new(*authn.Provider)),
		provideAuthResultHandler,
		wire.Bind(new(OAuthResultInteractionFlow), new(*interactionflows.AuthAPIFlow)),
	)
	return nil
}

func provideLinkHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	sp sso.Provider,
	ap LinkAuthnProvider,
	op sso.OAuthProvider,
) http.Handler {
	h := &LinkHandler{
		TxContext:     tx,
		Validator:     v,
		SSOProvider:   sp,
		AuthnProvider: ap,
		OAuthProvider: op,
	}
	return requireAuthz(h, h)
}

func newLinkHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authn.ProvideAuthAPIProvider,
		wire.Bind(new(LinkAuthnProvider), new(*authn.Provider)),
		provideOAuthProviderFromRequestVars,
		provideLinkHandler,
		ProvideRedirectURIForAPIFunc,
	)
	return nil
}

func provideLoginHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	sp sso.Provider,
	ap LoginAuthnProvider,
	op sso.OAuthProvider,
) http.Handler {
	h := &LoginHandler{
		TxContext:     tx,
		Validator:     v,
		SSOProvider:   sp,
		AuthnProvider: ap,
		OAuthProvider: op,
	}
	return requireAuthz(h, h)
}

func newLoginHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		authn.ProvideAuthAPIProvider,
		wire.Bind(new(LoginAuthnProvider), new(*authn.Provider)),
		provideOAuthProviderFromRequestVars,
		provideLoginHandler,
		ProvideRedirectURIForAPIFunc,
	)
	return nil
}

func provideAuthRedirectHandler(
	sp sso.Provider,
	op sso.OAuthProvider,
) http.Handler {
	h := &AuthRedirectHandler{
		SSOProvider:   sp,
		OAuthProvider: op,
	}
	return h
}

func newAuthRedirectHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		provideOAuthProviderFromRequestVars,
		provideAuthRedirectHandler,
		ProvideRedirectURIForAPIFunc,
	)
	return nil
}

func provideAuthURLHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	pp password.Provider,
	sp sso.Provider,
	cfg *config.TenantConfiguration,
	op sso.OAuthProvider,
	action ssoAction,
) http.Handler {
	h := &AuthURLHandler{
		TxContext:                  tx,
		Validator:                  v,
		PasswordAuthProvider:       pp,
		SSOProvider:                sp,
		OAuthConflictConfiguration: cfg.AppConfig.AuthAPI.OnIdentityConflict.OAuth,
		OAuthProvider:              op,
		Action:                     action,
	}
	return requireAuthz(h, h)
}

func providerLoginSSOAction() ssoAction {
	return ssoActionLogin
}

func newLoginAuthURLHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		provideOAuthProviderFromRequestVars,
		provideAuthURLHandler,
		ProvideRedirectURIForAPIFunc,
		providerLoginSSOAction,
	)
	return nil
}

func providerLinkSSOAction() ssoAction {
	return ssoActionLink
}

func newLinkAuthURLHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		provideOAuthProviderFromRequestVars,
		provideAuthURLHandler,
		ProvideRedirectURIForAPIFunc,
		providerLinkSSOAction,
	)
	return nil
}

func providerUnlinkHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	oap oauth.Provider,
	ais authinfo.Store,
	ups userprofile.Store,
	hp hook.Provider,
	spf *sso.OAuthProviderFactory,
) http.Handler {
	h := &UnlinkHandler{
		TxContext:         tx,
		OAuthAuthProvider: oap,
		AuthInfoStore:     ais,
		UserProfileStore:  ups,
		HookProvider:      hp,
		ProviderFactory:   spf,
	}
	return requireAuthz(h, h)
}

func newUnlinkHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		providerUnlinkHandler,
		ProvideRedirectURIForAPIFunc,
	)
	return nil
}
