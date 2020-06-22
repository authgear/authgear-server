package redis

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func ProvideStore(
	ctx context.Context,
	c *config.TenantConfiguration,
	t clock.Clock,
	lf logging.Factory,
) session.Store {
	return NewStore(ctx, c.AppID, t, lf)
}

var DependencySet = wire.NewSet(
	ProvideStore,
)
