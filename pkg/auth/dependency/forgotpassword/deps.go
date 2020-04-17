package forgotpassword

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(StoreImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	ProvideProvider,
)

func ProvideProvider(
	tConfig *config.TenantConfiguration,
	store Store,
	ais authinfo.Store,
	ups userprofile.Store,
	pp password.Provider,
	pc *audit.PasswordChecker,
	hp hook.Provider,
	tp coretime.Provider,
	upp urlprefix.Provider,
	te *template.Engine,
	ms mail.Sender,
	sc sms.Client,
	tq async.Queue,
) *Provider {
	return &Provider{
		AppName:                     tConfig.AppConfig.DisplayAppName,
		EmailMessageConfiguration:   tConfig.AppConfig.Messages.Email,
		ForgotPasswordConfiguration: tConfig.AppConfig.ForgotPassword,
		Store:                       store,
		AuthInfoStore:               ais,
		UserProfileStore:            ups,
		PasswordAuthProvider:        pp,
		PasswordChecker:             pc,
		HookProvider:                hp,
		TimeProvider:                tp,
		URLPrefixProvider:           upp,
		TemplateEngine:              te,
		MailSender:                  ms,
		SMSClient:                   sc,
		TaskQueue:                   tq,
	}
}
