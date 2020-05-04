package oob

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/template"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	c *config.TenantConfiguration,
	sqlb db.SQLBuilder,
	sqle db.SQLExecutor,
	t time.Provider,
	te *template.Engine,
	upp urlprefix.Provider,
	tq async.Queue,
) *Provider {
	return &Provider{
		Config:                    c.AppConfig.Authenticator.OOB,
		SMSMessageConfiguration:   c.AppConfig.Messages.SMS,
		EmailMessageConfiguration: c.AppConfig.Messages.Email,
		Store:                     &Store{SQLBuilder: sqlb, SQLExecutor: sqle},
		Time:                      t,
		TemplateEngine:            te,
		URLPrefixProvider:         upp,
		TaskQueue:                 tq,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
