package authflowclient

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewClient,
)
