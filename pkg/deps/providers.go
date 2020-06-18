package deps

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/log"
	"github.com/skygeario/skygear-server/pkg/redis"
	taskexecutors "github.com/skygeario/skygear-server/pkg/task/executors"
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

	loggerFactory := log.NewFactory(
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

func (p *RootProvider) NewRequestProvider(ctx context.Context, r *http.Request, cfg *config.Config) *RequestProvider {
	loggerFactory := p.LoggerFactory.WithHooks(log.NewSecretMaskLogHook(cfg.SecretConfig))
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	dbContext := db.NewContext(
		ctx,
		p.DatabasePool,
		cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials),
	)
	redisContext := redis.NewContext(
		p.RedisPool,
		cfg.SecretConfig.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials),
	)

	return &RequestProvider{
		RootProvider:  p,
		Request:       r,
		Context:       ctx,
		LoggerFactory: loggerFactory,
		Config:        cfg,
		DbContext:     dbContext,
		RedisContext:  redisContext,
	}
}

type RequestProvider struct {
	*RootProvider

	Request       *http.Request
	Context       context.Context
	LoggerFactory *log.Factory
	Config        *config.Config
	DbContext     db.Context
	RedisContext  *redis.Context
}
