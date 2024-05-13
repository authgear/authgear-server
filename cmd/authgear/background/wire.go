//go:build wireinject
// +build wireinject

package background

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/feature/accountanonymization"
	"github.com/authgear/authgear-server/pkg/lib/feature/accountdeletion"
	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
)

func newConfigSourceController(p *deps.BackgroundProvider, c context.Context) *configsource.Controller {
	panic(wire.Build(
		DependencySet,
		configsource.NewResolveAppIDTypeDomain,
		configsource.DependencySet,
		configsource.ControllerDependencySet,
	))
}

func newAccountAnonymizationRunner(p *deps.BackgroundProvider, c context.Context, ctrl *configsource.Controller) *backgroundjob.Runner {
	panic(wire.Build(
		DependencySet,
		accountanonymization.DependencySet,
		wire.Bind(new(accountanonymization.AppContextResolver), new(*configsource.Controller)),
	))
}

func newAccountDeletionRunner(p *deps.BackgroundProvider, c context.Context, ctrl *configsource.Controller) *backgroundjob.Runner {
	panic(wire.Build(
		DependencySet,
		accountdeletion.DependencySet,
		wire.Bind(new(accountdeletion.AppContextResolver), new(*configsource.Controller)),
	))
}

func newUserService(ctx context.Context, p *deps.BackgroundProvider, appID string, appContext *config.AppContext) *UserService {
	panic(wire.Build(
		DependencySet,
		wire.FieldsOf(new(*config.AppContext), "Config"),
	))
}
