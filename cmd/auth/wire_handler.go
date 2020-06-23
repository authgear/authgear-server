//+build wireinject

package main

import (
	"net/http"

	"github.com/google/wire"

	handleroauth "github.com/skygeario/skygear-server/pkg/auth/handler/oauth"
	handlersession "github.com/skygeario/skygear-server/pkg/auth/handler/session"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func newSessionResolveHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handlersession.ResolveHandler)),
	))
}

func newOAuthAuthorizeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.AuthorizeHandler)),
	))
}

func newOAuthTokenHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.TokenHandler)),
	))
}

func newOAuthRevokeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.RevokeHandler)),
	))
}

func newOAuthMetadataHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.MetadataHandler)),
	))
}

func newOAuthJWKSHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.JWKSHandler)),
	))
}

func newOAuthUserInfoHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.UserInfoHandler)),
	))
}

func newOAuthEndSessionHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.EndSessionHandler)),
	))
}

func newOAuthChallengeHandler(p *deps.RequestProvider) http.Handler {
	panic(wire.Build(
		deps.RequestDependencySet,
		wire.Bind(new(http.Handler), new(*handleroauth.ChallengeHandler)),
	))
}
