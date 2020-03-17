//+build wireinject

package sso

import (
	"net/http"

	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func provideOAuthProviderFromRequestVars(r *http.Request, spf *sso.OAuthProviderFactory) sso.OAuthProvider {
	vars := mux.Vars(r)
	return spf.NewOAuthProvider(vars["provider"])
}

func provideAuthHandler(
	tx db.TxContext,
	cfg *config.TenantConfiguration,
	hp sso.AuthHandlerHTMLProvider,
	sp sso.Provider,
	ap AuthHandlerAuthnProvider,
	op sso.OAuthProvider,
) http.Handler {
	h := &AuthHandler{
		TxContext:               tx,
		TenantConfiguration:     cfg,
		AuthHandlerHTMLProvider: hp,
		SSOProvider:             sp,
		AuthnProvider:           ap,
		OAuthProvider:           op,
	}
	return h
}

func newAuthHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(AuthHandlerAuthnProvider), new(*authn.Provider)),
		provideOAuthProviderFromRequestVars,
		provideAuthHandler,
	)
	return nil
}

func provideAuthResultHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	ap AuthResultAuthnProvider,
	v *validation.Validator,
	sp sso.Provider,
) http.Handler {
	h := &AuthResultHandler{
		TxContext:     tx,
		AuthnProvider: ap,
		Validator:     v,
		SSOProvider:   sp,
	}
	return requireAuthz(h, h)
}

func newAuthResultHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(AuthResultAuthnProvider), new(*authn.Provider)),
		provideAuthResultHandler,
	)
	return nil
}

func provideLinkHandler(
	tx db.TxContext,
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	ac coreauth.ContextGetter,
	sp sso.Provider,
	ap LinkAuthnProvider,
	op sso.OAuthProvider,
) http.Handler {
	h := &LinkHandler{
		TxContext:     tx,
		Validator:     v,
		AuthContext:   ac,
		SSOProvider:   sp,
		AuthnProvider: ap,
		OAuthProvider: op,
	}
	return requireAuthz(h, h)
}

func newLinkHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(LinkAuthnProvider), new(*authn.Provider)),
		provideOAuthProviderFromRequestVars,
		provideLinkHandler,
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

func newLoginHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(LoginAuthnProvider), new(*authn.Provider)),
		provideOAuthProviderFromRequestVars,
		provideLoginHandler,
	)
	return nil
}