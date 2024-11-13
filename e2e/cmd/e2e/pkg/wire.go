//go:build wireinject
// +build wireinject

package e2e

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
)

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
	panic(wire.Build(
		configsource.NewResolveAppIDTypeDomain,
		deps.RootDependencySet,
		configsource.ControllerDependencySet,
	))
}

func newUserImport(p *deps.AppProvider) *userimport.UserImportService {
	panic(wire.Build(
		End2EndDependencySet,
		deps.CommonDependencySet,
	))
}

func newLoginIDSerivce(p *deps.AppProvider) *loginid.Provider {
	panic(wire.Build(
		End2EndDependencySet,
		deps.CommonDependencySet,
	))
}
