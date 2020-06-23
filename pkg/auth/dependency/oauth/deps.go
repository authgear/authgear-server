package oauth

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(MetadataProvider), "*"),
	wire.Struct(new(Resolver), "*"),
	wire.Bind(new(auth.AccessTokenSessionResolver), new(*Resolver)),
	wire.Struct(new(SessionManager), "*"),
	wire.Bind(new(auth.AccessTokenSessionManager), new(*SessionManager)),
)
