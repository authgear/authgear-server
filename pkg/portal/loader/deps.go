package loader

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(ViewerLoader), "*"),
	wire.Struct(new(AppLoader), "*"),
	wire.Struct(new(DomainLoader), "*"),
)
