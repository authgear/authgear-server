package event

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
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
	request *http.Request,
	trustProxy config.TrustProxy,
	logger Logger,
	database Database,
	clock clock.Clock,
	users UserService,
	localization *config.LocalizationConfig,
	store Store,
	hookSink *hook.Sink,
	auditSink *audit.Sink,
) *Service {
	return &Service{
		Context:      ctx,
		Request:      request,
		TrustProxy:   trustProxy,
		Logger:       logger,
		Database:     database,
		Clock:        clock,
		Users:        users,
		Localization: localization,
		Store:        store,
		Sinks:        []Sink{hookSink, auditSink},
	}
}
