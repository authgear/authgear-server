package plan

import (
	configplan "github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	configplan.DependencySet,
	wire.Struct(new(Service), "*"),
)
