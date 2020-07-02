package deps

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	authtemplate "github.com/authgear/authgear-server/pkg/auth/template"
	"github.com/authgear/authgear-server/pkg/core/sentry"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/log"
	"github.com/authgear/authgear-server/pkg/redis"
	"github.com/authgear/authgear-server/pkg/task"
	taskexecutors "github.com/authgear/authgear-server/pkg/task/executors"
	"github.com/authgear/authgear-server/pkg/template"
)

type RootProvider struct {
	ServerConfig        *config.ServerConfig
	LoggerFactory       *log.Factory
	DatabasePool        *db.Pool
	RedisPool           *redis.Pool
	TaskExecutor        *taskexecutors.InMemoryExecutor
	ReservedNameChecker *loginid.ReservedNameChecker
}

func NewRootProvider(cfg *config.ServerConfig) (*RootProvider, error) {
	var p RootProvider

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	loggerFactory := log.NewFactory(
		logLevel,
		log.NewDefaultMaskLogHook(),
		&sentry.LogHook{Hub: sentry.DefaultClient.Hub},
	)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	taskExecutor := taskexecutors.NewInMemoryExecutor(loggerFactory, ProvideRestoreTaskContext(&p))
	reservedNameChecker, err := loginid.NewReservedNameChecker(cfg.ReservedNameFilePath)
	if err != nil {
		return nil, err
	}

	p = RootProvider{
		ServerConfig:        cfg,
		LoggerFactory:       loggerFactory,
		DatabasePool:        dbPool,
		RedisPool:           redisPool,
		TaskExecutor:        taskExecutor,
		ReservedNameChecker: reservedNameChecker,
	}
	return &p, nil
}

func (p *RootProvider) NewAppProvider(ctx context.Context, cfg *config.Config) *AppProvider {
	loggerFactory := p.LoggerFactory.WithHooks(log.NewSecretMaskLogHook(cfg.SecretConfig))
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
	templateEngine := authtemplate.NewEngineWithConfig(cfg)

	return &AppProvider{
		RootProvider:   p,
		Context:        ctx,
		Config:         cfg,
		LoggerFactory:  loggerFactory,
		Database:       database,
		Redis:          redis,
		TemplateEngine: templateEngine,
	}
}

func (p *RootProvider) NewRequestProvider(r *http.Request, cfg *config.Config) *RequestProvider {
	ap := p.NewAppProvider(r.Context(), cfg)

	return &RequestProvider{
		AppProvider: ap,
		Request:     r,
	}
}

func (p *RootProvider) NewTaskProvider(ctx context.Context, cfg *config.Config) *TaskProvider {
	ap := p.NewAppProvider(ctx, cfg)

	return &TaskProvider{
		AppProvider: ap,
	}
}

func (p *RootProvider) Handler(factory func(*RequestProvider) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := getRequestProvider(r.Context())
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
			p := getRequestProvider(r.Context())
			m := factory(p)
			h := m.Handle(next)
			h.ServeHTTP(w, r)
		})
	})
}

func (p *RootProvider) Task(factory func(provider *TaskProvider) task.Task) task.Task {
	return TaskFunc(func(ctx context.Context, param interface{}) error {
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
	Database       db.Context
	Redis          *redis.Handle
	TemplateEngine *template.Engine
}

type RequestProvider struct {
	*AppProvider

	Request *http.Request
}

type TaskProvider struct {
	*AppProvider
}
