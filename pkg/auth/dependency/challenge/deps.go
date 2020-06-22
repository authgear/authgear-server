package challenge

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideProvider(
	ctx context.Context,
	t clock.Clock,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Context: ctx,
		AppID:   c.AppID,
		Clock:   t,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
