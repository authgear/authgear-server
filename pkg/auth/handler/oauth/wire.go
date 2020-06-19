package oauth

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/deps"
)

func newAuthorizeHandler(p *deps.RequestProvider) http.Handler {
	return (*AuthorizeHandler)(nil)
}

func newTokenHandler(p *deps.RequestProvider) http.Handler {
	return (*TokenHandler)(nil)
}

func newRevokeHandler(p *deps.RequestProvider) http.Handler {
	return (*RevokeHandler)(nil)
}

func newMetadataHandler(p *deps.RequestProvider) http.Handler {
	return (*MetadataHandler)(nil)
}

func newJWKSHandler(p *deps.RequestProvider) http.Handler {
	return (*JWKSHandler)(nil)
}

func newUserInfoHandler(p *deps.RequestProvider) http.Handler {
	return (*UserInfoHandler)(nil)
}

func newEndSessionHandler(p *deps.RequestProvider) http.Handler {
	return (*EndSessionHandler)(nil)
}

func newChallengeHandler(p *deps.RequestProvider) http.Handler {
	return (*ChallengeHandler)(nil)
}
