package oauth

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(MetadataProvider), "*"),
	wire.Struct(new(Resolver), "*"),
	wire.Struct(new(SessionManager), "*"),
	wire.Struct(new(URLProvider), "*"),

	wire.Struct(new(AccessTokenEncoding), "*"),
	wire.Bind(new(AccessTokenDecoder), new(*AccessTokenEncoding)),
)
