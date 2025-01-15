package deps

import (
	"context"
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type RootProvider struct {
	EnvironmentConfig  *config.EnvironmentConfig
	ConfigSourceConfig *configsource.Config
	LoggerFactory      *log.Factory
	SentryHub          *getsentry.Hub
	DatabasePool       *db.Pool
	RedisPool          *redis.Pool
	RedisHub           *redis.Hub
	BaseResources      *resource.Manager
	EmbeddedResources  *web.GlobalEmbeddedResourceManager
}

func NewRootProvider(
	ctx context.Context,
	cfg *config.EnvironmentConfig,
	configSourceConfig *configsource.Config,
	customResourceDirectory string,
) (*RootProvider, error) {
	var p RootProvider

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	sentryHub, err := sentry.NewHub(string(cfg.SentryDSN))
	if err != nil {
		return nil, err
	}

	loggerFactory := log.NewFactory(
		logLevel,
		apierrors.SkipLoggingHook{},
		log.NewDefaultMaskLogHook(),
		sentry.NewLogHookFromHub(sentryHub),
	)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	redisHub := redis.NewHub(ctx, redisPool, loggerFactory)

	embeddedResources, err := web.NewDefaultGlobalEmbeddedResourceManager()
	if err != nil {
		return nil, err
	}

	p = RootProvider{
		EnvironmentConfig:  cfg,
		ConfigSourceConfig: configSourceConfig,
		LoggerFactory:      loggerFactory,
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
	return &p, nil
}

func (p *RootProvider) NewAppProvider(ctx context.Context, appCtx *config.AppContext) *AppProvider {
	cfg := appCtx.Config
	loggerFactory := p.LoggerFactory.ReplaceHooks(
		apierrors.SkipLoggingHook{},
		log.NewDefaultMaskLogHook(),
		config.NewSecretMaskLogHook(cfg.SecretConfig),
		// NewAppProvider is used in 2 places.
		// 1. Process normal incoming HTTP requests. In this case, sentry middleware will inject a more detailed sentry.Hub in the context.
		// 2. Process async tasks. In this case, there is no sentry middleware and the context is context.Background(), so we need to fallback to use p.SentryHub.
		sentry.NewLogHookFromContextOrFallback(ctx, p.SentryHub),
	)
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	appDatabase := appdb.NewHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials),
		loggerFactory,
	)
	var searchDatabaseCredentials *config.SearchDatabaseCredentials
	if s := cfg.SecretConfig.LookupData(config.SearchDatabaseCredentialsKey); s != nil {
		searchDatabaseCredentials = s.(*config.SearchDatabaseCredentials)
	}
	searchDatabase := searchdb.NewHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		searchDatabaseCredentials,
		loggerFactory,
	)
	var auditDatabaseCredentials *config.AuditDatabaseCredentials
	if a := cfg.SecretConfig.LookupData(config.AuditDatabaseCredentialsKey); a != nil {
		auditDatabaseCredentials = a.(*config.AuditDatabaseCredentials)
	}
	auditReadDatabase := auditdb.NewReadHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		auditDatabaseCredentials,
		loggerFactory,
	)
	auditWriteDatabase := auditdb.NewWriteHandle(
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		auditDatabaseCredentials,
		loggerFactory,
	)
	redis := appredis.NewHandle(
		p.RedisPool,
		p.RedisHub,
		&p.EnvironmentConfig.RedisConfig,
		cfg.SecretConfig.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials),
		loggerFactory,
	)
	globalRedis := globalredis.NewHandle(
		p.RedisPool,
		&p.EnvironmentConfig.RedisConfig,
		&p.EnvironmentConfig.GlobalRedis,
		loggerFactory,
	)

	var analyticRedisCredentials *config.AnalyticRedisCredentials
	if c := cfg.SecretConfig.LookupData(config.AnalyticRedisCredentialsKey); c != nil {
		analyticRedisCredentials = c.(*config.AnalyticRedisCredentials)
	}
	analyticRedis := analyticredis.NewHandle(
		p.RedisPool,
		&p.EnvironmentConfig.RedisConfig,
		analyticRedisCredentials,
		loggerFactory,
	)

	provider := &AppProvider{
		RootProvider:       p,
		LoggerFactory:      loggerFactory,
		AppDatabase:        appDatabase,
		SearchDatabase:     searchDatabase,
		AuditReadDatabase:  auditReadDatabase,
		AuditWriteDatabase: auditWriteDatabase,
		Redis:              redis,
		GlobalRedis:        globalRedis,
		AnalyticRedis:      analyticRedis,
		AppContext:         appCtx,
	}
	return provider
}

func (p *RootProvider) RootHandler(factory func(*RootProvider, http.ResponseWriter, *http.Request, context.Context) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := factory(p, w, r, r.Context())
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

	LoggerFactory      *log.Factory
	AppDatabase        *appdb.Handle
	SearchDatabase     *searchdb.Handle
	AuditReadDatabase  *auditdb.ReadHandle
	AuditWriteDatabase *auditdb.WriteHandle
	Redis              *appredis.Handle
	AnalyticRedis      *analyticredis.Handle
	AppContext         *config.AppContext
	GlobalRedis        *globalredis.Handle
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
	LoggerFactory      *log.Factory
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
) (*BackgroundProvider, error) {
	var p BackgroundProvider

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	sentryHub, err := sentry.NewHub(string(cfg.SentryDSN))
	if err != nil {
		return nil, err
	}

	loggerFactory := log.NewFactory(
		logLevel,
		apierrors.SkipLoggingHook{},
		log.NewDefaultMaskLogHook(),
		sentry.NewLogHookFromHub(sentryHub),
	)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	redisHub := redis.NewHub(ctx, redisPool, loggerFactory)

	embeddedResources, err := web.NewDefaultGlobalEmbeddedResourceManager()
	if err != nil {
		return nil, err
	}

	p = BackgroundProvider{
		EnvironmentConfig:  cfg,
		ConfigSourceConfig: configSourceConfig,
		LoggerFactory:      loggerFactory,
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

	return &p, nil
}
