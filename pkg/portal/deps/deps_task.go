package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	featureweb3 "github.com/authgear/authgear-server/pkg/lib/feature/web3"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/task"
	"github.com/authgear/authgear-server/pkg/portal/task/tasks"
)

func ProvideSMTPServerCredentials(c *portalconfig.SMTPConfig) *config.SMTPServerCredentials {
	return &config.SMTPServerCredentials{
		Host:     c.Host,
		Port:     c.Port,
		Username: c.Username,
		Password: c.Password,
		Mode:     c.Mode,
	}
}

var TaskDependencySet = wire.NewSet(
	ProvideSMTPServerCredentials,

	tasks.DependencySet,
	mail.DependencySet,
	sms.DependencySet,
	featureweb3.DependencySet,
	wire.Bind(new(tasks.MailSender), new(*mail.Sender)),
	wire.Bind(new(tasks.NFTService), new(*featureweb3.Service)),

	task.DependencySet,
)
