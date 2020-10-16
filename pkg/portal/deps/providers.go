package deps

import (
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type RootProvider struct {
	EnvironmentConfig        *config.EnvironmentConfig
	ConfigSourceConfig       *configsource.Config
	AuthgearConfig           *portalconfig.AuthgearConfig
	AdminAPIConfig           *portalconfig.AdminAPIConfig
	AppConfig                *portalconfig.AppConfig
	DatabaseConfig           *portalconfig.DatabaseConfig
	SMTPConfig               *portalconfig.SMTPConfig
	MailConfig               *portalconfig.MailConfig
	DefaultResourceDirectory portalconfig.DefaultResourceDirectory
	LoggerFactory            *log.Factory
	SentryHub                *getsentry.Hub

	ConfigSourceController *configsource.Controller
	Resources              *resource.Manager
}

func NewRootProvider(
	cfg *config.EnvironmentConfig,
	defaultResourceDirectory portalconfig.DefaultResourceDirectory,
	configSourceConfig *configsource.Config,
	authgearConfig *portalconfig.AuthgearConfig,
	adminAPIConfig *portalconfig.AdminAPIConfig,
	appConfig *portalconfig.AppConfig,
	dbConfig *portalconfig.DatabaseConfig,
	smtpConfig *portalconfig.SMTPConfig,
	mailConfig *portalconfig.MailConfig,
) (*RootProvider, error) {
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

	return &RootProvider{
		EnvironmentConfig:        cfg,
		DefaultResourceDirectory: defaultResourceDirectory,
		ConfigSourceConfig:       configSourceConfig,
		AuthgearConfig:           authgearConfig,
		AdminAPIConfig:           adminAPIConfig,
		AppConfig:                appConfig,
		DatabaseConfig:           dbConfig,
		SMTPConfig:               smtpConfig,
		MailConfig:               mailConfig,
		LoggerFactory:            loggerFactory,
		SentryHub:                sentryHub,
		Resources:                NewResourceManager(string(defaultResourceDirectory)),
	}, nil
}

type RequestProvider struct {
	RootProvider *RootProvider
	Request      *http.Request
}

func (p *RootProvider) Middleware(f func(*RequestProvider) httproute.Middleware) httproute.Middleware {
	return httproute.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestProvider := &RequestProvider{
				RootProvider: p,
				Request:      r,
			}
			m := f(requestProvider)
			h := m.Handle(next)
			h.ServeHTTP(w, r)
		})
	})
}

func (p *RootProvider) Handler(f func(*RequestProvider) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestProvider := &RequestProvider{
			RootProvider: p,
			Request:      r,
		}
		h := f(requestProvider)
		h.ServeHTTP(w, r)
	})
}
