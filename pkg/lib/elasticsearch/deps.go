package elasticsearch

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewClient,
	NewElasticsearchServiceLogger,
	wire.Struct(new(Service), "*"),
)
