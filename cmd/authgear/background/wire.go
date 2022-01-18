//go:build wireinject
// +build wireinject

package background

import (
	"context"
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
)

func newConfigSourceController(p *deps.BackgroundProvider, c context.Context) *configsource.Controller {
	panic(wire.Build(
		deps.BackgroundDependencySet,
	))
}
