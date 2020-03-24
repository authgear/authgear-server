//+build wireinject

package oauth

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func provideAuthorizeHandler(lf logging.Factory, tx db.TxContext, ah oauthAuthorizeHandler) http.Handler {
	h := &AuthorizeHandler{
		logger:       lf.NewLogger("oauth-authz-handler"),
		txContext:    tx,
		authzHandler: ah,
	}
	return h
}

func newAuthorizeHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(oauthAuthorizeHandler), new(*handler.AuthorizationHandler)),
		provideAuthorizeHandler,
	)
	return nil
}

func provideTokenHandler(lf logging.Factory, tx db.TxContext, th oauthTokenHandler) http.Handler {
	h := &TokenHandler{
		logger:       lf.NewLogger("oauth-token-handler"),
		txContext:    tx,
		tokenHandler: th,
	}
	return h
}

func newTokenHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(oauthTokenHandler), new(*handler.TokenHandler)),
		wire.Bind(new(handler.IDTokenIssuer), new(*oidc.IDTokenIssuer)),
		provideTokenHandler,
	)
	return nil
}

func provideRevokeHandler(lf logging.Factory, tx db.TxContext, rh oauthRevokeHandler) http.Handler {
	h := &RevokeHandler{
		logger:        lf.NewLogger("oauth-revoke-handler"),
		txContext:     tx,
		revokeHandler: rh,
	}
	return h
}

func newRevokeHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(oauthRevokeHandler), new(*handler.RevokeHandler)),
		provideRevokeHandler,
	)
	return nil
}

func provideMetadataHandler(oauth *oauth.MetadataProvider, oidc *oidc.MetadataProvider) http.Handler {
	h := &MetadataHandler{
		metaProviders: []oauthMetadataProvider{oauth, oidc},
	}
	return h
}

func newMetadataHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		provideMetadataHandler,
	)
	return nil
}

func provideJWKSHandler(config *config.TenantConfiguration) http.Handler {
	h := &JWKSHandler{
		config: *config.AppConfig.OIDC,
	}
	return h
}

func newJWKSHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		provideJWKSHandler,
	)
	return nil
}

func provideUserInfoHandler(lf logging.Factory, tx db.TxContext, uip oauthUserInfoProvider) http.Handler {
	h := &UserInfoHandler{
		logger:           lf.NewLogger("oauth-userinfo-handler"),
		txContext:        tx,
		userInfoProvider: uip,
	}
	return h
}

func newUserInfoHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	wire.Build(
		auth.DependencySet,
		wire.Bind(new(oauthUserInfoProvider), new(*oidc.IDTokenIssuer)),
		provideUserInfoHandler,
	)
	return nil
}
