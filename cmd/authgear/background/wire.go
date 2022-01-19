//go:build wireinject
// +build wireinject

package background

import (
	"context"
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/feature/accountdeletion"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
)

func newConfigSourceController(p *deps.BackgroundProvider, c context.Context) *configsource.Controller {
	panic(wire.Build(
		deps.BackgroundDependencySet,
	))
}

func newAccountDeletionRunner(p *deps.BackgroundProvider, c context.Context) *backgroundjob.Runner {
	panic(wire.Build(
		deps.BackgroundDependencySet,
		accountdeletion.DependencySet,
	))
}
