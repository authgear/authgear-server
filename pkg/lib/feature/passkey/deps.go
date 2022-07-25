package passkey

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(ConfigService), "*"),
	wire.Struct(new(CreationOptionsService), "*"),
	wire.Struct(new(Service), "*"),
	wire.Struct(new(Store), "*"),
)
