package challenge

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideProvider(
	ctx context.Context,
	t time.Provider,
	c *config.TenantConfiguration,
) *Provider {
	return &Provider{
		Context: ctx,
		AppID:   c.AppID,
		Time:    t,
	}
}

var DependencySet = wire.NewSet(ProvideProvider)
