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
)
