package images

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/images/deps"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
)
