package reindex

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(SourceProvider), "*"),
	NewReindexerLogger,
	wire.Struct(new(Reindexer), "*"),
	NewSinkLogger,
	wire.Struct(new(Sink), "*"),
)
