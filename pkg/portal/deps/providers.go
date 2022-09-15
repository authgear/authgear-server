package deps

import (
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type RootProvider struct {
	EnvironmentConfig      *config.EnvironmentConfig
	ConfigSourceConfig     *configsource.Config
	AuthgearConfig         *portalconfig.AuthgearConfig
	AdminAPIConfig         *portalconfig.AdminAPIConfig
	AppConfig              *portalconfig.AppConfig
	SMTPConfig             *portalconfig.SMTPConfig
	MailConfig             *portalconfig.MailConfig
	KubernetesConfig       *portalconfig.KubernetesConfig
	DomainImplementation   portalconfig.DomainImplementationType
	SearchConfig           *portalconfig.SearchConfig
	Web3Config             *portalconfig.Web3Config
	AuditLogConfig         *portalconfig.AuditLogConfig
	AnalyticConfig         *config.AnalyticConfig
	StripeConfig           *portalconfig.StripeConfig
	GoogleTagManagerConfig *portalconfig.GoogleTagManagerConfig
	LoggerFactory          *log.Factory
	SentryHub              *getsentry.Hub

	Database               *db.Pool
	RedisPool              *redis.Pool
	GlobalRedisHandle      *globalredis.Handle
	ConfigSourceController *configsource.Controller
	Resources              *resource.Manager
	AppBaseResources       *resource.Manager
	FilesystemCache        *httputil.FilesystemCache
}

func NewRootProvider(
	cfg *config.EnvironmentConfig,
	builtinResourceDirectory string,
	customResourceDirectory string,
	appBuiltinResourceDirectory string,
	appCustomResourceDirectory string,
	configSourceConfig *configsource.Config,
	authgearConfig *portalconfig.AuthgearConfig,
	adminAPIConfig *portalconfig.AdminAPIConfig,
	appConfig *portalconfig.AppConfig,
	smtpConfig *portalconfig.SMTPConfig,
	mailConfig *portalconfig.MailConfig,
	kubernetesConfig *portalconfig.KubernetesConfig,
	domainImplementation portalconfig.DomainImplementationType,
	searchConfig *portalconfig.SearchConfig,
	web3Config *portalconfig.Web3Config,
	auditLogConfig *portalconfig.AuditLogConfig,
	analyticConfig *config.AnalyticConfig,
	stripeConfig *portalconfig.StripeConfig,
	googleTagManagerConfig *portalconfig.GoogleTagManagerConfig,
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

	redisPool := redis.NewPool()
	globalRedisHandle := globalredis.NewHandle(
		redisPool,
		&cfg.RedisConfig,
		&cfg.GlobalRedis,
		loggerFactory,
	)

	filesystemCache := httputil.NewFilesystemCache()

	return &RootProvider{
		EnvironmentConfig:      cfg,
		ConfigSourceConfig:     configSourceConfig,
		AuthgearConfig:         authgearConfig,
		AdminAPIConfig:         adminAPIConfig,
		AppConfig:              appConfig,
		SMTPConfig:             smtpConfig,
		MailConfig:             mailConfig,
		KubernetesConfig:       kubernetesConfig,
		DomainImplementation:   domainImplementation,
		SearchConfig:           searchConfig,
		Web3Config:             web3Config,
		AuditLogConfig:         auditLogConfig,
		AnalyticConfig:         analyticConfig,
		StripeConfig:           stripeConfig,
		GoogleTagManagerConfig: googleTagManagerConfig,
		LoggerFactory:          loggerFactory,
		SentryHub:              sentryHub,
		Database:               db.NewPool(),
		RedisPool:              redisPool,
		GlobalRedisHandle:      globalRedisHandle,
		Resources: resource.NewManagerWithDir(
			portalresource.PortalRegistry,
			builtinResourceDirectory,
			customResourceDirectory,
		),
		AppBaseResources: resource.NewManagerWithDir(
			resource.DefaultRegistry,
			appBuiltinResourceDirectory,
			appCustomResourceDirectory,
		),
		FilesystemCache: filesystemCache,
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
