package portal

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/portal/loader"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/transport"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	service.DependencySet,
	wire.Bind(new(service.ConfigGetter), new(*deps.ConfigGetter)),

	loader.DependencySet,
	wire.Bind(new(loader.AppService), new(*service.AppService)),

	graphql.DependencySet,
	wire.Bind(new(graphql.ViewerLoader), new(*loader.ViewerLoader)),
	wire.Bind(new(graphql.AppLoader), new(*loader.AppLoader)),

	transport.DependencySet,
)
