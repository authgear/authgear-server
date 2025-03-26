package globalredis

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewHandle,
)
