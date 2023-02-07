package universallink

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(IOSAssociatedDomainsHandler), "*"),
	wire.Struct(new(AndroidAssociatedDomainsHandler), "*"),
)
