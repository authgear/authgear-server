package reindex

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewReindexerLogger,
	wire.Struct(new(Reindexer), "*"),
)
