package deps

import (
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type RootProvider struct {
	EnvironmentConfig    *config.EnvironmentConfig
	ConfigSourceConfig   *configsource.Config
	AuthgearConfig       *portalconfig.AuthgearConfig
	AdminAPIConfig       *portalconfig.AdminAPIConfig
	AppConfig            *portalconfig.AppConfig
	DatabaseConfig       *config.DatabaseEnvironmentConfig
	SMTPConfig           *portalconfig.SMTPConfig
	MailConfig           *portalconfig.MailConfig
	KubernetesConfig     *portalconfig.KubernetesConfig
	DomainImplementation portalconfig.DomainImplementationType
	SearchConfig         *portalconfig.SearchConfig
	AuditLogConfig       *portalconfig.AuditLogConfig
	LoggerFactory        *log.Factory
	SentryHub            *getsentry.Hub

	Database               *db.Pool
	ConfigSourceController *configsource.Controller
	Resources              *resource.Manager
	AppBaseResources       *resource.Manager

	SecretKeyAllowlist portalconfig.SecretKeyAllowlist
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
	dbConfig *config.DatabaseEnvironmentConfig,
	smtpConfig *portalconfig.SMTPConfig,
	mailConfig *portalconfig.MailConfig,
	kubernetesConfig *portalconfig.KubernetesConfig,
	domainImplementation portalconfig.DomainImplementationType,
	searchConfig *portalconfig.SearchConfig,
	auditLogConfig *portalconfig.AuditLogConfig,
	secretKeyAllowlist portalconfig.SecretKeyAllowlist,
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
		EnvironmentConfig:    cfg,
		ConfigSourceConfig:   configSourceConfig,
		AuthgearConfig:       authgearConfig,
		AdminAPIConfig:       adminAPIConfig,
		AppConfig:            appConfig,
		DatabaseConfig:       dbConfig,
		SMTPConfig:           smtpConfig,
		MailConfig:           mailConfig,
		KubernetesConfig:     kubernetesConfig,
		DomainImplementation: domainImplementation,
		SearchConfig:         searchConfig,
		AuditLogConfig:       auditLogConfig,
		LoggerFactory:        loggerFactory,
		SentryHub:            sentryHub,
		Database:             db.NewPool(),
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
		SecretKeyAllowlist: secretKeyAllowlist,
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
