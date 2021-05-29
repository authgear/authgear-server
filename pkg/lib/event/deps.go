package event

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var DependencySet = wire.NewSet(
	NewLogger,
	NewService,
	wire.Struct(new(StoreImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
)

func NewService(
	ctx context.Context,
	logger Logger,
	database Database,
	clock clock.Clock,
	users UserService,
	localization *config.LocalizationConfig,
	store Store,
	hookSink *hook.Sink,
) *Service {
	return &Service{
		Context:      ctx,
		Logger:       logger,
		Database:     database,
		Clock:        clock,
		Users:        users,
		Localization: localization,
		Store:        store,
		Sinks:        []Sink{hookSink},
	}
}
