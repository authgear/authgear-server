package deps

import (
	getsentry "github.com/getsentry/sentry-go"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type RootProvider struct {
	EnvironmentConfig imagesconfig.EnvironmentConfig
	LoggerFactory     *log.Factory
	SentryHub         *getsentry.Hub
}

func NewRootProvider(
	cfg imagesconfig.EnvironmentConfig,
) (*RootProvider, error) {
	logLevel, err := log.ParseLevel(string(cfg.LogLevel))
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

	return &RootProvider{
		EnvironmentConfig: cfg,
		LoggerFactory:     loggerFactory,
		SentryHub:         sentryHub,
	}, nil
}
