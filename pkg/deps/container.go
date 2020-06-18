package deps

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/log"
	"github.com/skygeario/skygear-server/pkg/redis"
)

type RootContainer struct {
	ServerConfig        *config.ServerConfig
	LoggerFactory       *log.Factory
	DatabasePool        db.Pool
	RedisPool           *redis.Pool
	AsyncTaskExecutor   *async.Executor
	ReservedNameChecker *loginid.ReservedNameChecker
}

func NewRootContainer(cfg *config.ServerConfig) (*RootContainer, error) {
	loggerFactory := log.NewFactory(
		log.NewDefaultLogHook(),
		&sentry.LogHook{Hub: sentry.DefaultClient.Hub},
	)
	dbPool := db.NewPool()
	redisPool := redis.NewPool()
	asyncTaskExecutor := async.NewExecutor(dbPool)
	reservedNameChecker, err := loginid.NewReservedNameChecker(cfg.ReservedNameFilePath)
	if err != nil {
		return nil, err
	}

	return &RootContainer{
		ServerConfig:        cfg,
		LoggerFactory:       loggerFactory,
		DatabasePool:        dbPool,
		RedisPool:           redisPool,
		AsyncTaskExecutor:   asyncTaskExecutor,
		ReservedNameChecker: reservedNameChecker,
	}, nil
}

func (c *RootContainer) NewRequestContainer(cfg *config.Config) *RequestContainer {
	loggerFactory := c.LoggerFactory.WithHooks(log.NewSecretLogHook(cfg.SecretConfig))
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID

	return &RequestContainer{
		RootContainer: c,
		LoggerFactory: loggerFactory,
		Config:        cfg,
	}
}

type RequestContainer struct {
	*RootContainer

	LoggerFactory *log.Factory
	Config        *config.Config
}
