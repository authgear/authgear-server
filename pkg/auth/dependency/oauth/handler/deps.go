package handler

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
)

var DependencySet = wire.NewSet(
	NewAuthorizationHandlerLogger,
	wire.Struct(new(AuthorizationHandler), "*"),
	NewTokenHandlerLogger,
	wire.Struct(new(TokenHandler), "*"),
	wire.Struct(new(RevokeHandler), "*"),
	wire.Value(TokenGenerator(oauth.GenerateToken)),
	wire.Struct(new(oauth.URLProvider), "*"),
)
