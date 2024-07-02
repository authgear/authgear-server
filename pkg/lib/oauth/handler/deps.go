package handler

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewAuthorizationHandlerLogger,
	wire.Struct(new(AuthorizationHandler), "*"),
	NewTokenHandlerLogger,
	wire.Struct(new(TokenHandler), "*"),
	wire.Struct(new(RevokeHandler), "*"),
	NewAnonymousUserHandlerLogger,
	wire.Struct(new(AnonymousUserHandler), "*"),
	wire.Struct(new(TokenService), "*"),
	wire.Struct(new(CodeGrantService), "*"),
	wire.Struct(new(SettingsActionGrantService), "*"),
	wire.Struct(new(ProxyRedirectHandler), "*"),
	wire.Bind(new(TokenHandlerTokenService), new(*TokenService)),
)
