package user

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type Provider struct {
	*Commands
	*Queries
}

func ProvideCommands(
	us store,
	tp time.Provider,
	hp hook.Provider,
	up urlprefix.Provider,
	q async.Queue,
	config *config.TenantConfiguration,
	wmp WelcomeMessageProvider,
) *Commands {
	return &Commands{
		Store:                         us,
		Time:                          tp,
		Hooks:                         hp,
		URLPrefix:                     up,
		TaskQueue:                     q,
		UserVerificationConfiguration: config.AppConfig.UserVerification,
		WelcomeMessageProvider:        wmp,
	}
}

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Bind(new(store), new(*Store)),
	ProvideCommands,
	wire.Struct(new(Queries), "*"),
	wire.Struct(new(Provider), "*"),
)
