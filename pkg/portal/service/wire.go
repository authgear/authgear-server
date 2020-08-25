//+build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/deps"
)

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
	panic(wire.Build(deps.DependencySet))
}
