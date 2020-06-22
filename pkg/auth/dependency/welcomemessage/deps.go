package welcomemessage

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/template"
)

func ProvideProvider(
	ctx context.Context,
	c *config.TenantConfiguration,
	templateEngine *template.Engine,
	taskQueue async.Queue,
) *Provider {
	return &Provider{
		Context:                   ctx,
		LocalizationConfiguration: c.AppConfig.Localization,
		MetadataConfiguration:     c.AppConfig.AuthUI.Metadata,
		EmailConfig: config.NewEmailMessageConfiguration(
			c.AppConfig.Messages.Email,
			c.AppConfig.WelcomeMessage.EmailMessage,
		),
		WelcomeMessageConfiguration: c.AppConfig.WelcomeMessage,
		TemplateEngine:              templateEngine,
		TaskQueue:                   taskQueue,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
