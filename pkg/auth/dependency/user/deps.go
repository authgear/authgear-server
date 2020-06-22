package user

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Provider struct {
	*Commands
	*Queries
}

func ProvideRawCommands(
	us store,
	tp clock.Clock,
	up urlprefix.Provider,
	q async.Queue,
	config *config.TenantConfiguration,
	wmp WelcomeMessageProvider,
) *RawCommands {
	return &RawCommands{
		Store:                         us,
		Clock:                         tp,
		URLPrefix:                     up,
		TaskQueue:                     q,
		UserVerificationConfiguration: config.AppConfig.UserVerification,
		WelcomeMessageProvider:        wmp,
	}
}

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Bind(new(store), new(*Store)),
	wire.Struct(new(Commands), "*"),
	ProvideRawCommands,
	wire.Struct(new(Queries), "*"),
	wire.Struct(new(Provider), "*"),
)
