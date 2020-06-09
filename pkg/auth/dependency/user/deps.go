package user

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type Provider struct {
	*Commands
	*Queries
}

func ProvideRawCommands(
	us store,
	tp time.Provider,
	up urlprefix.Provider,
	q async.Queue,
	config *config.TenantConfiguration,
	wmp WelcomeMessageProvider,
) *RawCommands {
	return &RawCommands{
		Store:                         us,
		Time:                          tp,
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
