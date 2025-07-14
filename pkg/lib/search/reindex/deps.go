package reindex

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(SourceProvider), "*"),
	wire.Struct(new(Reindexer), "*"),
	wire.Struct(new(Sink), "*"),
)
