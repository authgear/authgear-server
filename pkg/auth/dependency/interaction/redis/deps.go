package redis

import (
	"context"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideStore(ctx context.Context, config *config.TenantConfiguration, time clock.Clock) *Store {
	return &Store{Context: ctx, AppID: config.AppID, Clock: time}
}

var DependencySet = wire.NewSet(
	ProvideStore,
	wire.Bind(new(interaction.Store), new(*Store)),
)
