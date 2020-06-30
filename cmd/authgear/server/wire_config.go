//+build wireinject

package server

import (
	"github.com/google/wire"

	configsource "github.com/authgear/authgear-server/pkg/auth/config/source"
	"github.com/authgear/authgear-server/pkg/deps"
)

func newConfigSource(p *deps.RootProvider) configsource.Source {
	panic(wire.Build(deps.RootDependencySet))
}
