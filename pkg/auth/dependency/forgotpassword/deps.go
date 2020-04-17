package forgotpassword

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
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
	pp password.Provider,
	store Store,
	tp coretime.Provider,
	upp urlprefix.Provider,
	te *template.Engine,
	ms mail.Sender,
	sc sms.Client,
) *Provider {
	return &Provider{
		AppName:                     tConfig.AppConfig.DisplayAppName,
		EmailMessageConfiguration:   tConfig.AppConfig.Messages.Email,
		ForgotPasswordConfiguration: tConfig.AppConfig.ForgotPassword,
		PasswordAuthProvider:        pp,
		Store:                       store,
		TimeProvider:                tp,
		URLPrefixProvider:           upp,
		TemplateEngine:              te,
		MailSender:                  ms,
		SMSClient:                   sc,
	}
}
