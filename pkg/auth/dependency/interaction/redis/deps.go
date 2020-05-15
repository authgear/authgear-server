package redis

import (
	"context"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideStore(ctx context.Context, config *config.TenantConfiguration, time time.Provider) *Store {
	return &Store{Context: ctx, AppID: config.AppID, Time: time}
}

var DependencySet = wire.NewSet(
	ProvideStore,
	wire.Bind(new(interaction.Store), new(*Store)),
)
