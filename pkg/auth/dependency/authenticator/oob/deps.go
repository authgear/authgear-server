package oob

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func ProvideProvider(
	ctx context.Context,
	c *config.TenantConfiguration,
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t clock.Clock,
	te *template.Engine,
	upp urlprefix.Provider,
	tq async.Queue,
) *Provider {
	return &Provider{
		Context:                   ctx,
		LocalizationConfiguration: c.AppConfig.Localization,
		MetadataConfiguration:     c.AppConfig.AuthUI.Metadata,
		Config:                    c.AppConfig.Authenticator.OOB,
		SMSMessageConfiguration:   c.AppConfig.Messages.SMS,
		EmailMessageConfiguration: c.AppConfig.Messages.Email,
		Store:                     &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Clock:                     t,
		TemplateEngine:            te,
		URLPrefixProvider:         upp,
		TaskQueue:                 tq,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
