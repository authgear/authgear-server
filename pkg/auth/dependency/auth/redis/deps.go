package redis

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideEventStore(
	ctx context.Context,
	c *config.TenantConfiguration,
) *EventStore {
	return NewEventStore(ctx, c.AppID)
}

var DependencySet = wire.NewSet(
	ProvideEventStore,
	wire.Bind(new(auth.AccessEventStore), new(*EventStore)),
)
