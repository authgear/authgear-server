package event

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
	"github.com/authgear/authgear-server/pkg/lib/userinfo"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var DependencySet = wire.NewSet(
	NewService,
	NewStoreImpl,
	wire.Struct(new(ResolverImpl), "*"),
	wire.Bind(new(Store), new(*StoreImpl)),
	wire.Bind(new(Resolver), new(*ResolverImpl)),
)

func NewService(
	appID config.AppID,
	remoteIP httputil.RemoteIP,
	userAgentString httputil.UserAgentString,
	httpRequestURL httputil.HTTPRequestURL,
	database Database,
	clock clock.Clock,
	localization *config.LocalizationConfig,
	store Store,
	resolver Resolver,
	hookSink *hook.Sink,
	auditSink *audit.Sink,
	searchSink *reindex.Sink,
	userInfoSink *userinfo.Sink,
) *Service {
	return &Service{
		AppID:           appID,
		RemoteIP:        remoteIP,
		UserAgentString: userAgentString,
		HTTPRequestURL:  httpRequestURL,
		Database:        database,
		Clock:           clock,
		Localization:    localization,
		Store:           store,
		Resolver:        resolver,
		Sinks: []Sink{
			hookSink,
			auditSink,
			searchSink,
			userInfoSink,
		},
	}
}
