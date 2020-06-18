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

type RootContainer struct {
	ServerConfig        *config.ServerConfig
	LoggerFactory       *log.Factory
	DatabasePool        *db.Pool
	RedisPool           *redis.Pool
	TaskExecutor        *taskexecutors.InMemoryExecutor
	ReservedNameChecker *loginid.ReservedNameChecker
}

func NewRootContainer(cfg *config.ServerConfig) (*RootContainer, error) {
	var container RootContainer

	loggerFactory := log.NewFactory(
		log.NewDefaultMaskLogHook(),
		&sentry.LogHook{Hub: sentry.DefaultClient.Hub},
	)
	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	taskExecutor := taskexecutors.NewInMemoryExecutor(loggerFactory, ProvideRestoreTaskContext(&container))
	reservedNameChecker, err := loginid.NewReservedNameChecker(cfg.ReservedNameFilePath)
	if err != nil {
		return nil, err
	}

	container = RootContainer{
		ServerConfig:        cfg,
		LoggerFactory:       loggerFactory,
		DatabasePool:        dbPool,
		RedisPool:           redisPool,
		TaskExecutor:        taskExecutor,
		ReservedNameChecker: reservedNameChecker,
	}
	return &container, nil
}

func (c *RootContainer) NewRequestContainer(ctx context.Context, r *http.Request, cfg *config.Config) *RequestContainer {
	loggerFactory := c.LoggerFactory.WithHooks(log.NewSecretMaskLogHook(cfg.SecretConfig))
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	dbContext := db.NewContext(
		ctx,
		c.DatabasePool,
		cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials),
	)
	redisContext := redis.NewContext(
		c.RedisPool,
		cfg.SecretConfig.LookupData(config.RedisCredentialsKey).(*config.RedisCredentials),
	)

	return &RequestContainer{
		RootContainer: c,
		Request:       r,
		Context:       ctx,
		LoggerFactory: loggerFactory,
		Config:        cfg,
		DbContext:     dbContext,
		RedisContext:  redisContext,
	}
}

type RequestContainer struct {
	*RootContainer

	Request       *http.Request
	Context       context.Context
	LoggerFactory *log.Factory
	Config        *config.Config
	DbContext     db.Context
	RedisContext  *redis.Context
}
