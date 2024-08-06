package ldap

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Client), "*"),
	wire.Struct(new(Service), "*"),
)
