package api

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(PersonService), "*"),
	NewClient,
	wire.Bind(new(PersonHTTPClient), new(*Client)),
)
