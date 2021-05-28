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
)

func NewService(
	ctx context.Context,
	logger Logger,
	database Database,
	clock clock.Clock,
	users UserService,
	localization *config.LocalizationConfig,
	hookSink *hook.Sink,
) *Service {
	return &Service{
		Context:      ctx,
		Logger:       logger,
		Database:     database,
		Clock:        clock,
		Users:        users,
		Localization: localization,
		Sinks:        []Sink{hookSink},
	}
}
