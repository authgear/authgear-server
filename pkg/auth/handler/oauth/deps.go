package oauth

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AuthorizeHandler), "*"),
	wire.Struct(new(ConsentHandler), "*"),
	wire.Struct(new(TokenHandler), "*"),
	wire.Struct(new(RevokeHandler), "*"),
	wire.Struct(new(MetadataHandler), "*"),
	wire.Struct(new(JWKSHandler), "*"),
	wire.Struct(new(UserInfoHandler), "*"),
	wire.Struct(new(EndSessionHandler), "*"),
	wire.Struct(new(ChallengeHandler), "*"),
	wire.Struct(new(AppSessionTokenHandler), "*"),
	wire.Struct(new(ProxyRedirectHandler), "*"),
)
