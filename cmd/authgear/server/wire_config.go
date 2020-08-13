//+build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/deps"
	configsource "github.com/authgear/authgear-server/pkg/lib/config/source"
)

func newConfigSource(p *deps.RootProvider) configsource.Source {
	panic(wire.Build(deps.RootDependencySet))
}
