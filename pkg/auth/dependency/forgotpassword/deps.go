package forgotpassword

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/deps"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(StoreImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	ProvideProvider,
)

func ProvideProvider(
	ctx context.Context,
	saup deps.StaticAssetURLPrefix,
	tConfig *config.TenantConfiguration,
	store Store,
	u UserProvider,
	hp hook.Provider,
	tp coretime.Provider,
	upp urlprefix.Provider,
	te *template.Engine,
	tq async.Queue,
	f ResetPasswordFlow,
	ip LoginIDProvider,
) *Provider {
	return &Provider{
		Context:                     ctx,
		LocalizationConfiguration:   tConfig.AppConfig.Localization,
		MetadataConfiguration:       tConfig.AppConfig.AuthUI.Metadata,
		StaticAssetURLPrefix:        string(saup),
		EmailMessageConfiguration:   tConfig.AppConfig.Messages.Email,
		SMSMessageConfiguration:     tConfig.AppConfig.Messages.SMS,
		ForgotPasswordConfiguration: tConfig.AppConfig.ForgotPassword,
		Store:                       store,
		Users:                       u,
		HookProvider:                hp,
		TimeProvider:                tp,
		URLPrefixProvider:           upp,
		TemplateEngine:              te,
		TaskQueue:                   tq,
		Interactions:                f,
		LoginIDProvider:             ip,
	}
}
