package event

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var DependencySet = wire.NewSet(
	NewLogger,
	NewService,
	wire.Struct(new(StoreImpl), "*"),
	wire.Struct(new(ResolverImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	wire.Bind(new(Resolver), new(*ResolverImpl)),
)

func NewService(
	ctx context.Context,
	remoteIP httputil.RemoteIP,
	userAgentString httputil.UserAgentString,
	logger Logger,
	database Database,
	clock clock.Clock,
	localization *config.LocalizationConfig,
	store Store,
	resolver Resolver,
	hookSink *hook.Sink,
	auditSink *audit.Sink,
	tutorialSink *tutorial.Sink,
	elasticSearchSink *elasticsearch.Sink,
) *Service {
	return &Service{
		Context:         ctx,
		RemoteIP:        remoteIP,
		UserAgentString: userAgentString,
		Logger:          logger,
		Database:        database,
		Clock:           clock,
		Localization:    localization,
		Store:           store,
		Resolver:        resolver,
		Sinks:           []Sink{hookSink, auditSink, elasticSearchSink, tutorialSink},
	}
}
