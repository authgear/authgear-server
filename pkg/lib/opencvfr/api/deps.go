package api

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(PersonService), "*"),
	wire.Struct(new(CollectionService), "*"),
	wire.Struct(new(SearchService), "*"),
	NewClient,
	wire.Bind(new(PersonHTTPClient), new(*Client)),
	wire.Bind(new(CollectionHTTPClient), new(*Client)),
	wire.Bind(new(SearchHTTPClient), new(*Client)),
)
