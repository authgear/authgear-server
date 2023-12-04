package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/task"
	"github.com/authgear/authgear-server/pkg/portal/task/tasks"
)

func ProvideTestModeEmailSuppressed() config.FeatureTestModeEmailSuppressed {
	return config.FeatureTestModeEmailSuppressed(false)
}

func ProvideSMTPServerCredentials(c *portalconfig.SMTPConfig) *config.SMTPServerCredentials {
	return &config.SMTPServerCredentials{
		Host:     c.Host,
		Port:     c.Port,
		Username: c.Username,
		Password: c.Password,
		Mode:     c.Mode,
	}
}

func ProvideEmptyTestModeConfig() *config.TestModeConfig {
	// We don't need test mode for portal
	c := &config.TestModeConfig{}
	config.SetFieldDefaults(c)
	return c
}

var testModeConfigDependencySet = wire.NewSet(
	ProvideEmptyTestModeConfig,
	wire.FieldsOf(new(*config.TestModeConfig),
		"FixedOOBOTP",
		"SMS",
		"Whatsapp",
		"Email",
	),
)

var TaskDependencySet = wire.NewSet(
	ProvideSMTPServerCredentials,
	testModeConfigDependencySet,

	tasks.DependencySet,
	mail.DependencySet,
	ProvideTestModeEmailSuppressed,
	wire.Bind(new(tasks.MailSender), new(*mail.Sender)),

	task.DependencySet,
)
