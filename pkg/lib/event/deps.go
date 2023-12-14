package event

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var DependencySet = wire.NewSet(
	NewLogger,
	NewService,
	NewStoreImpl,
	wire.Struct(new(ResolverImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	wire.Bind(new(Resolver), new(*ResolverImpl)),
)

func NewService(
	ctx context.Context,
	appID config.AppID,
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
	elasticSearchSink *elasticsearch.Sink,
	pgSearchSink *pgsearch.Sink,
) *Service {
	return &Service{
		Context:         ctx,
		AppID:           appID,
		RemoteIP:        remoteIP,
		UserAgentString: userAgentString,
		Logger:          logger,
		Database:        database,
		Clock:           clock,
		Localization:    localization,
		Store:           store,
		Resolver:        resolver,
		Sinks: []Sink{
			hookSink,
			auditSink,
			elasticSearchSink,
			pgSearchSink,
		},
	}
}
