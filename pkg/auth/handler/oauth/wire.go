//+build wireinject

package oauth

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
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
