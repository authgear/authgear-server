package sms

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideSMSClient(config *config.TenantConfiguration) Client {
	return NewClient(config.AppConfig)
}

var DependencySet = wire.NewSet(ProvideSMSClient)
