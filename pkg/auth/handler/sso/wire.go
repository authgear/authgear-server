//+build wireinject

package sso

import (
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"
	pkg "github.com/skygeario/skygear-server/pkg/auth"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
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
	op sso.OAuthProvider,
	f OAuthHandlerInteractionFlow,
) http.Handler {
	h := &AuthHandler{
		TxContext:               tx,
		TenantConfiguration:     cfg,
		AuthHandlerHTMLProvider: hp,
		SSOProvider:             sp,
		OAuthProvider:           op,
		Interactions:            f,
	}
	return h
}

func newAuthHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
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
	v *validation.Validator,
	sp sso.Provider,
	f OAuthResultInteractionFlow,
) http.Handler {
	h := &AuthResultHandler{
		TxContext:    tx,
		Validator:    v,
		SSOProvider:  sp,
		Interactions: f,
	}
	return requireAuthz(h, h)
}

func newAuthResultHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
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
	op sso.OAuthProvider,
	f OAuthLinkInteractionFlow,
) http.Handler {
	h := &LinkHandler{
		TxContext:     tx,
		Validator:     v,
		SSOProvider:   sp,
		OAuthProvider: op,
		Interactions:  f,
	}
	return requireAuthz(h, h)
}

func newLinkHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		provideOAuthProviderFromRequestVars,
		provideLinkHandler,
		ProvideRedirectURIForAPIFunc,
		wire.Bind(new(OAuthLinkInteractionFlow), new(*interactionflows.AuthAPIFlow)),
	)
	return nil
}

func provideLoginHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	sp sso.Provider,
	f OAuthLoginInteractionFlow,
	op sso.OAuthProvider,
) http.Handler {
	h := &LoginHandler{
		TxContext:     tx,
		Validator:     v,
		SSOProvider:   sp,
		OAuthProvider: op,
		Interactions:  f,
	}
	return requireAuthz(h, h)
}

func newLoginHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		provideOAuthProviderFromRequestVars,
		provideLoginHandler,
		ProvideRedirectURIForAPIFunc,
		wire.Bind(new(OAuthLoginInteractionFlow), new(*interactionflows.AuthAPIFlow)),
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
	sp sso.Provider,
	cfg *config.TenantConfiguration,
	op sso.OAuthProvider,
	action ssoAction,
) http.Handler {
	h := &AuthURLHandler{
		TxContext:                  tx,
		Validator:                  v,
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
	spf *sso.OAuthProviderFactory,
	f OAuthUnlinkInteractionFlow,
) http.Handler {
	h := &UnlinkHandler{
		TxContext:       tx,
		ProviderFactory: spf,
		Interactions:    f,
	}
	return requireAuthz(h, h)
}

func newUnlinkHandler(r *http.Request, m pkg.DependencyMap) http.Handler {
	wire.Build(
		pkg.DependencySet,
		providerUnlinkHandler,
		ProvideRedirectURIForAPIFunc,
		wire.Bind(new(OAuthUnlinkInteractionFlow), new(*interactionflows.AuthAPIFlow)),
	)
	return nil
}
