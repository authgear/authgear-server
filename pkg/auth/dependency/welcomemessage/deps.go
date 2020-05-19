package welcomemessage

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func ProvideProvider(
	c *config.TenantConfiguration,
	templateEngine *template.Engine,
	taskQueue async.Queue,
) *Provider {
	return &Provider{
		AppName: c.AppConfig.DisplayAppName,
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
