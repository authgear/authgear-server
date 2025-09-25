package handler

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AuthorizationHandler), "*"),
	wire.Struct(new(TokenHandler), "*"),
	wire.Struct(new(RevokeHandler), "*"),
	wire.Struct(new(AnonymousUserHandler), "*"),
	wire.Struct(new(TokenService), "*"),
	wire.Struct(new(CodeGrantService), "*"),
	wire.Struct(new(SettingsActionGrantService), "*"),
	wire.Struct(new(PreAuthenticatedURLTokenServiceImpl), "*"),
	wire.Bind(new(PreAuthenticatedURLTokenService), new(*PreAuthenticatedURLTokenServiceImpl)),
	wire.Struct(new(ProxyRedirectHandler), "*"),
	wire.Bind(new(TokenHandlerTokenService), new(*TokenService)),
	wire.Bind(new(TokenHandlerCodeGrantService), new(*CodeGrantService)),
)
