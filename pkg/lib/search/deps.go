package search

import (
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	reindex.DependencySet,
	elasticsearch.DependencySet,
	pgsearch.DependencySet,
	wire.Struct(new(Service), "*"),
)
