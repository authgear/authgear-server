package oauth

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
)

func ProvideResolverProvider(p session.Provider) ResolverSessionProvider { return p }

var DependencySet = wire.NewSet(
	wire.Struct(new(MetadataProvider), "*"),
	wire.Struct(new(Resolver), "*"),
	ProvideResolverProvider,
	wire.Bind(new(auth.AccessTokenSessionResolver), new(*Resolver)),
	wire.Struct(new(SessionManager), "*"),
	wire.Bind(new(auth.AccessTokenSessionManager), new(*SessionManager)),
)
