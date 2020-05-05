package mail

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideMailSender(ctx context.Context, config *config.TenantConfiguration) Sender {
	return NewSender(ctx, config)
}

var DependencySet = wire.NewSet(ProvideMailSender)
