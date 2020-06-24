//+build wireinject

package main

import (
	"github.com/google/wire"

	configsource "github.com/skygeario/skygear-server/pkg/auth/config/source"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func newConfigSource(p *deps.RootProvider) configsource.Source {
	panic(wire.Build(deps.RootDependencySet))
}
