package user

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type Provider struct {
	*Commands
	*Queries
}

func ProvideCommands(
	ais authinfo.Store,
	ups userprofile.Store,
	tp time.Provider,
	hp hook.Provider,
	up urlprefix.Provider,
	q async.Queue,
	config *config.TenantConfiguration,
	wmp WelcomeMessageProvider,
) *Commands {
	return &Commands{
		AuthInfos:                     ais,
		UserProfiles:                  ups,
		Time:                          tp,
		Hooks:                         hp,
		URLPrefix:                     up,
		TaskQueue:                     q,
		UserVerificationConfiguration: config.AppConfig.UserVerification,
		WelcomeMessageProvider:        wmp,
	}
}

var DependencySet = wire.NewSet(
	ProvideCommands,
	wire.Struct(new(Queries), "*"),
	wire.Struct(new(Provider), "*"),
)
