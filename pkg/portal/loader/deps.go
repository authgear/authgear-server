package loader

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewUserLoader,
	wire.Struct(new(AppLoader), "*"),
	wire.Struct(new(DomainLoader), "*"),
	wire.Struct(new(CollaboratorLoader), "*"),
)
