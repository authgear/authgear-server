package resourcescope

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Queries), "*"),
	wire.Struct(new(Commands), "*"),
	wire.Struct(new(ClientResourceScopeService), "*"),
)
