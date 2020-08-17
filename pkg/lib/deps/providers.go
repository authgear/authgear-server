package deps

import (
	"context"
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/sentry"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type RootProvider struct {
	ServerConfig        *config.ServerConfig
	LoggerFactory       *log.Factory
	SentryHub           *getsentry.Hub
	DatabasePool        *db.Pool
	RedisPool           *redis.Pool
	TaskQueueFactory    TaskQueueFactory
	ReservedNameChecker *loginid.ReservedNameChecker
}

func NewRootProvider(cfg *config.ServerConfig, taskQueueFactory TaskQueueFactory) (*RootProvider, error) {
	var p RootProvider

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	sentryHub, err := sentry.NewHub(cfg.SentryDSN)
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
	reservedNameChecker, err := loginid.NewReservedNameChecker(cfg.ReservedNameFilePath)
	if err != nil {
		return nil, err
	}

	p = RootProvider{
		ServerConfig:        cfg,
		LoggerFactory:       loggerFactory,
		SentryHub:           sentryHub,
		DatabasePool:        dbPool,
		RedisPool:           redisPool,
		TaskQueueFactory:    taskQueueFactory,
		ReservedNameChecker: reservedNameChecker,
	}
	return &p, nil
}

func (p *RootProvider) NewAppProvider(ctx context.Context, cfg *config.Config) *AppProvider {
	loggerFactory := p.LoggerFactory.ReplaceHooks(
		log.NewDefaultMaskLogHook(),
		log.NewSecretMaskLogHook(cfg.SecretConfig),
		sentry.NewLogHookFromContext(ctx),
	)
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	database := db.NewHandle(
		ctx,
		p.DatabasePool,
		cfg.AppConfig.Database,
		cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials),
		loggerFactory,
	)
	redis := redis.NewHandle(
		p.RedisPool,
		cfg.AppConfig.Redis,
		cfg.SecretConfig.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials),
		loggerFactory,
	)
	templateEngine := NewEngineWithConfig(p.ServerConfig, cfg)

	provider := &AppProvider{
		RootProvider:   p,
		Context:        ctx,
		Config:         cfg,
		LoggerFactory:  loggerFactory,
		Database:       database,
		Redis:          redis,
		TemplateEngine: templateEngine,
	}
	provider.TaskQueue = p.TaskQueueFactory(provider)
	return provider
}

func (p *RootProvider) Handler(factory func(*RequestProvider) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := getRequestProvider(r)
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
			p := getRequestProvider(r)
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

	Context        context.Context
	Config         *config.Config
	LoggerFactory  *log.Factory
	Database       *db.Handle
	Redis          *redis.Handle
	TaskQueue      task.Queue
	TemplateEngine *template.Engine
}

func (p *AppProvider) NewRequestProvider(r *http.Request) *RequestProvider {
	return &RequestProvider{
		AppProvider: p,
		Request:     r,
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

	Request *http.Request
}

type TaskProvider struct {
	*AppProvider

	Context context.Context
}
