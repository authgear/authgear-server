package oauthclient

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Resolver), "*"),
	wire.Struct(new(Store), "*"),
	wire.Struct(new(ClientCache), "*"),
	wire.Struct(new(Commands), "*"),
	wire.Struct(new(Queries), "*"),
)
