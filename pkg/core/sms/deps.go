package sms

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideSMSClient(ctx context.Context, config *config.TenantConfiguration) Client {
	return NewClient(ctx, config.AppConfig)
}

var DependencySet = wire.NewSet(ProvideSMSClient)
