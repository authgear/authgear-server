package deps

import (
	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type RootProvider struct {
	ServerConfig  *config.ServerConfig
	LoggerFactory *log.Factory
	SentryHub     *getsentry.Hub
}

func NewRootProvider(cfg *config.ServerConfig) (*RootProvider, error) {
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

	return &RootProvider{
		ServerConfig:  cfg,
		LoggerFactory: loggerFactory,
		SentryHub:     sentryHub,
	}, nil
}
