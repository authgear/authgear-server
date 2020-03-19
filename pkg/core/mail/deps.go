package mail

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideMailSender(config *config.TenantConfiguration) Sender {
	return NewSender(config.AppConfig.SMTP)
}

var DependencySet = wire.NewSet(ProvideMailSender)
