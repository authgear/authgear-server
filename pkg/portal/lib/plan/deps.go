package plan

import (
	"github.com/google/wire"

	configplan "github.com/authgear/authgear-server/pkg/lib/config/plan"
)

var DependencySet = wire.NewSet(
	configplan.DependencySet,
	wire.Struct(new(Service), "*"),
)
