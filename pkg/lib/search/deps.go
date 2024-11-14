package search

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
)

var DependencySet = wire.NewSet(
	reindex.DependencySet,
	elasticsearch.DependencySet,
	pgsearch.DependencySet,
	wire.Struct(new(Service), "*"),
)
