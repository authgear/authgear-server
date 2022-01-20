package deps

import (
	"context"
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
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
	TaskQueueFactory   TaskQueueFactory
	BaseResources      *resource.Manager
}

func NewRootProvider(
	cfg *config.EnvironmentConfig,
	configSourceConfig *configsource.Config,
	builtinResourceDirectory string,
	customResourceDirectory string,
	taskQueueFactory TaskQueueFactory,
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
		log.NewDefaultMaskLogHook(),
		sentry.NewLogHookFromHub(sentryHub),
	)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	redisHub := redis.NewHub(redisPool, loggerFactory)

	p = RootProvider{
		EnvironmentConfig:  cfg,
		ConfigSourceConfig: configSourceConfig,
		LoggerFactory:      loggerFactory,
		SentryHub:          sentryHub,
		DatabasePool:       dbPool,
		RedisPool:          redisPool,
		RedisHub:           redisHub,
		TaskQueueFactory:   taskQueueFactory,
		BaseResources: resource.NewManagerWithDir(
			resource.DefaultRegistry,
			builtinResourceDirectory,
			customResourceDirectory,
		),
	}
	return &p, nil
}

func (p *RootProvider) NewAppProvider(ctx context.Context, appCtx *config.AppContext) *AppProvider {
	cfg := appCtx.Config
	loggerFactory := p.LoggerFactory.ReplaceHooks(
		log.NewDefaultMaskLogHook(),
		config.NewSecretMaskLogHook(cfg.SecretConfig),
		sentry.NewLogHookFromContext(ctx),
	)
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	appDatabase := appdb.NewHandle(
		ctx,
		p.DatabasePool,
		cfg.AppConfig.Database,
		cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials),
		loggerFactory,
	)
	var auditDatabaseCredentials *config.AuditDatabaseCredentials
	if a := cfg.SecretConfig.LookupData(config.AuditDatabaseCredentialsKey); a != nil {
		auditDatabaseCredentials = a.(*config.AuditDatabaseCredentials)
	}
	auditReadDatabase := auditdb.NewReadHandle(
		ctx,
		p.DatabasePool,
		cfg.AppConfig.Database,
		auditDatabaseCredentials,
		loggerFactory,
	)
	auditWriteDatabase := auditdb.NewWriteHandle(
		ctx,
		p.DatabasePool,
		cfg.AppConfig.Database,
		auditDatabaseCredentials,
		loggerFactory,
	)
	redis := appredis.NewHandle(
		p.RedisPool,
		p.RedisHub,
		cfg.AppConfig.Redis,
		cfg.SecretConfig.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials),
		loggerFactory,
	)

	var analyticRedisCredentials *config.AnalyticRedisCredentials
	if c := cfg.SecretConfig.LookupData(config.AnalyticRedisCredentialsKey); c != nil {
		analyticRedisCredentials = c.(*config.AnalyticRedisCredentials)
	}
	analyticRedis := analyticredis.NewHandle(
		p.RedisPool,
		cfg.AppConfig.Redis,
		analyticRedisCredentials,
		loggerFactory,
	)

	provider := &AppProvider{
		RootProvider:       p,
		Context:            ctx,
		Config:             cfg,
		LoggerFactory:      loggerFactory,
		AppDatabase:        appDatabase,
		AuditReadDatabase:  auditReadDatabase,
		AuditWriteDatabase: auditWriteDatabase,
		Redis:              redis,
		AnalyticRedis:      analyticRedis,
		Resources:          appCtx.Resources,
	}
	provider.TaskQueue = p.TaskQueueFactory(provider)
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

func (p *RootProvider) Task(factory func(provider *TaskProvider) task.Task) task.Task {
	return TaskFunc(func(ctx context.Context, param task.Param) error {
		p := getTaskProvider(ctx)
		task := factory(p)
		return task.Run(ctx, param)
	})
}

type AppProvider struct {
	*RootProvider

	Context            context.Context
	Config             *config.Config
	LoggerFactory      *log.Factory
	AppDatabase        *appdb.Handle
	AuditReadDatabase  *auditdb.ReadHandle
	AuditWriteDatabase *auditdb.WriteHandle
	Redis              *appredis.Handle
	AnalyticRedis      *analyticredis.Handle
	TaskQueue          task.Queue
	Resources          *resource.Manager
}

func (p *AppProvider) NewRequestProvider(w http.ResponseWriter, r *http.Request) *RequestProvider {
	return &RequestProvider{
		AppProvider:    p,
		Request:        r,
		ResponseWriter: w,
	}
}

func (p *AppProvider) NewTaskProvider(ctx context.Context) *TaskProvider {
	return &TaskProvider{
		AppProvider: p,
		Context:     ctx,
	}
}

type RequestProvider struct {
	*AppProvider

	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

type TaskProvider struct {
	*AppProvider

	Context context.Context
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
}

func NewBackgroundProvider(
	cfg *config.EnvironmentConfig,
	configSourceConfig *configsource.Config,
	builtinResourceDirectory string,
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
		log.NewDefaultMaskLogHook(),
		sentry.NewLogHookFromHub(sentryHub),
	)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	redisHub := redis.NewHub(redisPool, loggerFactory)

	p = BackgroundProvider{
		EnvironmentConfig:  cfg,
		ConfigSourceConfig: configSourceConfig,
		LoggerFactory:      loggerFactory,
		SentryHub:          sentryHub,
		DatabasePool:       dbPool,
		RedisPool:          redisPool,
		RedisHub:           redisHub,
		BaseResources: resource.NewManagerWithDir(
			resource.DefaultRegistry,
			builtinResourceDirectory,
			customResourceDirectory,
		),
	}

	return &p, nil
}
