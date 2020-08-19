package portal

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/portal/loader"
	"github.com/authgear/authgear-server/pkg/portal/transport"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	loader.DependencySet,

	graphql.DependencySet,
	wire.Bind(new(graphql.ViewerLoader), new(*loader.ViewerLoader)),

	transport.DependencySet,
)
