package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/task"
	"github.com/authgear/authgear-server/pkg/worker/tasks"
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

// We do not send SMS for now.
func ProvideTwilioCredentials() *config.TwilioCredentials {
	return nil
}

// We do not send SMS for now.
func ProvideNexmoCrednetials() *config.NexmoCredentials {
	return nil
}

// We do not send SMS for now.
func ProvideMessageConfig() *config.MessagingConfig {
	return &config.MessagingConfig{}
}

var TaskDependencySet = wire.NewSet(
	ProvideSMTPServerCredentials,
	ProvideTwilioCredentials,
	ProvideNexmoCrednetials,
	ProvideMessageConfig,

	tasks.DependencySet,
	mail.DependencySet,
	sms.DependencySet,
	wire.Bind(new(tasks.MailSender), new(*mail.Sender)),
	wire.Bind(new(tasks.SMSClient), new(*sms.Client)),

	task.DependencySet,
)
