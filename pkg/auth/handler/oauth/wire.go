//+build wireinject

package oauth

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
)

func provideAuthorizeHandler(ah oauthAuthorizeHandler) http.Handler {
	h := &AuthorizeHandler{
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
