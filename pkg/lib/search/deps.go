package search

import (
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	elasticsearch.DependencySet,
	pgsearch.DependencySet,
	wire.Struct(new(Service), "*"),
)
