package flows

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideUserController(
	ais authinfo.Store,
	ups userprofile.Store,
	ti AuthAPITokenIssuer,
	scc session.CookieConfiguration,
	sp session.Provider,
	hp hook.Provider,
	tp time.Provider,
	c *config.TenantConfiguration,
) *UserController {
	return &UserController{
		AuthInfos:           ais,
		UserProfiles:        ups,
		TokenIssuer:         ti,
		SessionCookieConfig: scc,
		Sessions:            sp,
		Hooks:               hp,
		Time:                tp,
		Clients:             c.AppConfig.Clients,
	}
}

var DependencySet = wire.NewSet(
	wire.Struct(new(WebAppFlow), "*"),
	wire.Struct(new(AuthAPIFlow), "*"),
	wire.Struct(new(AnonymousFlow), "*"),
	wire.Struct(new(PasswordFlow), "*"),
	ProvideUserController,
)
