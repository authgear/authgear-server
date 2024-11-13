package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
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

var MailDependencySet = wire.NewSet(
	ProvideSMTPServerCredentials,
	mail.DependencySet,
)
