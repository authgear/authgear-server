package healthz

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewHandlerLogger,
	wire.Struct(new(Handler), "*"),
)
