package elasticsearch

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewClient,
	NewElasticsearchServiceLogger,
	wire.Struct(new(Service), "*"),
	NewLogger,
	wire.Struct(new(Sink), "*"),
)
