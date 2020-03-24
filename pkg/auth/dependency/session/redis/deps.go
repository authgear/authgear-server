package redis

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideStore(
	ctx context.Context,
	c *config.TenantConfiguration,
	t time.Provider,
	lf logging.Factory,
) session.Store {
	return NewStore(ctx, c.AppID, t, lf)
}

var DependencySet = wire.NewSet(
	ProvideStore,
)
