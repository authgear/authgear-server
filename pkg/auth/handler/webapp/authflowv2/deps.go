package authflowv2

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AuthflowV2LoginHandler), "*"),
)
