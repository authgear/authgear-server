package user

import (
	"github.com/google/wire"
)

type Provider struct {
	*Commands
	*Queries
}

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Bind(new(store), new(*Store)),
	wire.Struct(new(Commands), "*"),
	wire.Struct(new(Queries), "*"),
	wire.Struct(new(Provider), "*"),
	wire.Struct(new(RawCommands), "*"),
)
