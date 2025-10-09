package deps

import (
	"context"
	"net/http"

	getsentry "github.com/getsentry/sentry-go"
	"github.com/lestrrat-go/jwx/v2/jwk"

	runtimeresource "github.com/authgear/authgear-server"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type RootProvider struct {
	EnvironmentConfig  *config.EnvironmentConfig
	ConfigSourceConfig *configsource.Config
	SentryHub          *getsentry.Hub
	DatabasePool       *db.Pool
	RedisPool          *redis.Pool
	RedisHub           *redis.Hub
	BaseResources      *resource.Manager
	EmbeddedResources  *web.GlobalEmbeddedResourceManager
	JWKCache           *jwk.Cache
}

func NewRootProvider(
	ctx context.Context,
	cfg *config.EnvironmentConfig,
	configSourceConfig *configsource.Config,
	customResourceDirectory string,
) (context.Context, *RootProvider, error) {
	var p RootProvider

	sentryHub, err := sentry.NewHub(string(cfg.SentryDSN))
	if err != nil {
		return ctx, nil, err
	}
	ctx = getsentry.SetHubOnContext(ctx, sentryHub)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	redisHub := redis.NewHub(ctx, redisPool)

	embeddedResources, err := web.NewDefaultGlobalEmbeddedResourceManager()
	if err != nil {
		return ctx, nil, err
	}

	jwkCache := jwk.NewCache(ctx)

	p = RootProvider{
		EnvironmentConfig:  cfg,
		ConfigSourceConfig: configSourceConfig,
		SentryHub:          sentryHub,
		DatabasePool:       dbPool,
		RedisPool:          redisPool,
		RedisHub:           redisHub,
		BaseResources: resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
			Registry:              resource.DefaultRegistry,
			BuiltinResourceFS:     runtimeresource.EmbedFS_resources_authgear,
			BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_authgear,
			CustomResourceDir:     customResourceDirectory,
		}),
		EmbeddedResources: embeddedResources,
		JWKCache:          jwkCache,
	}
	return ctx, &p, nil
}

func (p *RootProvider) NewAppProvider(ctx context.Context, appCtx *config.AppContext) (context.Context, *AppProvider) {
	cfg := appCtx.Config

	appDatabase := appdb.NewHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials),
	)
	var searchDatabaseCredentials *config.SearchDatabaseCredentials
	if s := cfg.SecretConfig.LookupData(config.SearchDatabaseCredentialsKey); s != nil {
		searchDatabaseCredentials = s.(*config.SearchDatabaseCredentials)
	}
	searchDatabase := searchdb.NewHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		searchDatabaseCredentials,
	)
	var auditDatabaseCredentials *config.AuditDatabaseCredentials
	if a := cfg.SecretConfig.LookupData(config.AuditDatabaseCredentialsKey); a != nil {
		auditDatabaseCredentials = a.(*config.AuditDatabaseCredentials)
	}
	auditReadDatabase := auditdb.NewReadHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		auditDatabaseCredentials,
	)
	auditWriteDatabase := auditdb.NewWriteHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		auditDatabaseCredentials,
	)
	redis := appredis.NewHandle(
		p.RedisPool,
		p.RedisHub,
		&p.EnvironmentConfig.RedisConfig,
		cfg.SecretConfig.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials),
	)

	var analyticRedisCredentials *config.AnalyticRedisCredentials
	if c := cfg.SecretConfig.LookupData(config.AnalyticRedisCredentialsKey); c != nil {
		analyticRedisCredentials = c.(*config.AnalyticRedisCredentials)
	}
	analyticRedis := analyticredis.NewHandle(
		p.RedisPool,
		&p.EnvironmentConfig.RedisConfig,
		analyticRedisCredentials,
	)

	provider := &AppProvider{
		RootProvider:       p,
		AppDatabase:        appDatabase,
		SearchDatabase:     searchDatabase,
		AuditReadDatabase:  auditReadDatabase,
		AuditWriteDatabase: auditWriteDatabase,
		Redis:              redis,
		AnalyticRedis:      analyticRedis,
		AppContext:         appCtx,
	}
	return ctx, provider
}

func (p *RootProvider) RootHandler(factory func(*RootProvider, http.ResponseWriter, *http.Request, context.Context) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := factory(p, w, r, r.Context())
		h.ServeHTTP(w, r)
	})
}

func (p *RootProvider) RootHandlerWithConfigSource(cfgSource *configsource.ConfigSource, factory func(*configsource.ConfigSource, *RootProvider, http.ResponseWriter, *http.Request, context.Context) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := factory(cfgSource, p, w, r, r.Context())
		h.ServeHTTP(w, r)
	})
}

func (p *RootProvider) Handler(factory func(*RequestProvider) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := getRequestProvider(w, r)
		h := factory(p)
		h.ServeHTTP(w, r)
	})
}

func (p *RootProvider) RootMiddleware(factory func(*RootProvider) httproute.Middleware) httproute.Middleware {
	return factory(p)
}

func (p *RootProvider) Middleware(factory func(*RequestProvider) httproute.Middleware) httproute.Middleware {
	return httproute.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := getRequestProvider(w, r)
			m := factory(p)
			h := m.Handle(next)
			h.ServeHTTP(w, r)
		})
	})
}

type AppProvider struct {
	*RootProvider

	AppDatabase        *appdb.Handle
	SearchDatabase     *searchdb.Handle
	AuditReadDatabase  *auditdb.ReadHandle
	AuditWriteDatabase *auditdb.WriteHandle
	Redis              *appredis.Handle
	AnalyticRedis      *analyticredis.Handle
	AppContext         *config.AppContext
}

func (p *AppProvider) NewRequestProvider(w http.ResponseWriter, r *http.Request) *RequestProvider {
	return &RequestProvider{
		AppProvider:    p,
		Request:        r,
		ResponseWriter: w,
	}
}

type RequestProvider struct {
	*AppProvider

	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

type BackgroundProvider struct {
	EnvironmentConfig  *config.EnvironmentConfig
	ConfigSourceConfig *configsource.Config
	SentryHub          *getsentry.Hub
	DatabasePool       *db.Pool
	RedisPool          *redis.Pool
	RedisHub           *redis.Hub
	BaseResources      *resource.Manager
	EmbeddedResources  *web.GlobalEmbeddedResourceManager
}

func NewBackgroundProvider(
	ctx context.Context,
	cfg *config.EnvironmentConfig,
	configSourceConfig *configsource.Config,
	customResourceDirectory string,
) (context.Context, *BackgroundProvider, error) {
	var p BackgroundProvider

	sentryHub, err := sentry.NewHub(string(cfg.SentryDSN))
	if err != nil {
		return ctx, nil, err
	}
	ctx = getsentry.SetHubOnContext(ctx, sentryHub)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	redisHub := redis.NewHub(ctx, redisPool)

	embeddedResources, err := web.NewDefaultGlobalEmbeddedResourceManager()
	if err != nil {
		return ctx, nil, err
	}

	p = BackgroundProvider{
		EnvironmentConfig:  cfg,
		ConfigSourceConfig: configSourceConfig,
		SentryHub:          sentryHub,
		DatabasePool:       dbPool,
		RedisPool:          redisPool,
		RedisHub:           redisHub,
		BaseResources: resource.NewManagerWithDir(resource.NewManagerWithDirOptions{
			Registry:              resource.DefaultRegistry,
			BuiltinResourceFS:     runtimeresource.EmbedFS_resources_authgear,
			BuiltinResourceFSRoot: runtimeresource.RelativePath_resources_authgear,
			CustomResourceDir:     customResourceDirectory,
		}),
		EmbeddedResources: embeddedResources,
	}

	return ctx, &p, nil
}
