package oauth

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewAuthorizeHandlerLogger,
	wire.Struct(new(AuthorizeHandler), "*"),
	NewTokenHandlerLogger,
	wire.Struct(new(TokenHandler), "*"),
	NewRevokeHandlerLogger,
	wire.Struct(new(RevokeHandler), "*"),
	wire.Struct(new(MetadataHandler), "*"),
	NewJWKSHandlerLogger,
	wire.Struct(new(JWKSHandler), "*"),
	NewUserInfoHandlerLogger,
	wire.Struct(new(UserInfoHandler), "*"),
	NewEndSessionHandlerLogger,
	wire.Struct(new(EndSessionHandler), "*"),
	wire.Struct(new(ChallengeHandler), "*"),
)
