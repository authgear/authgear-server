package deps

import (
	"context"
	"net/http"
	"runtime"

	getsentry "github.com/getsentry/sentry-go"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/sentry"
	"github.com/authgear/authgear-server/pkg/util/vipsutil"
)

type RootProvider struct {
	EnvironmentConfig imagesconfig.EnvironmentConfig
	ObjectStoreConfig *imagesconfig.ObjectStoreConfig
	LoggerFactory     *log.Factory
	SentryHub         *getsentry.Hub
	DatabasePool      *db.Pool
	VipsDaemon        *vipsutil.Daemon
	BaseResources     *resource.Manager
}

func NewRootProvider(
	envConfig imagesconfig.EnvironmentConfig,
	objectStoreConfig *imagesconfig.ObjectStoreConfig,
) (*RootProvider, error) {
	logLevel, err := log.ParseLevel(string(envConfig.LogLevel))
	if err != nil {
		return nil, err
	}

	sentryHub, err := sentry.NewHub(string(envConfig.SentryDSN))
	if err != nil {
		return nil, err
	}

	loggerFactory := log.NewFactory(
		logLevel,
		log.NewDefaultMaskLogHook(),
		sentry.NewLogHookFromHub(sentryHub),
	)

	dbPool := db.NewPool()

	// We do not have a chance to close this yet :(
	// But it is not harmful not to close this.
	vipsDaemon := vipsutil.OpenDaemon(runtime.NumCPU())

	return &RootProvider{
		EnvironmentConfig: envConfig,
		ObjectStoreConfig: objectStoreConfig,
		LoggerFactory:     loggerFactory,
		SentryHub:         sentryHub,
		DatabasePool:      dbPool,
		VipsDaemon:        vipsDaemon,
		BaseResources: resource.NewManagerWithDir(
			resource.DefaultRegistry,
			envConfig.BuiltinResourceDirectory,
			envConfig.CustomResourceDirectory,
		),
	}, nil
}

func (p *RootProvider) NewAppProvider(ctx context.Context, appCtx *config.AppContext) *AppProvider {
	cfg := appCtx.Config
	loggerFactory := p.LoggerFactory.ReplaceHooks(
		log.NewDefaultMaskLogHook(),
		config.NewSecretMaskLogHook(cfg.SecretConfig),
		sentry.NewLogHookFromContext(ctx),
	)
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	provider := &AppProvider{
		RootProvider:  p,
		Context:       ctx,
		Config:        cfg,
		LoggerFactory: loggerFactory,
	}
	return provider
}

func (p *RootProvider) RootMiddleware(factory func(*RootProvider) httproute.Middleware) httproute.Middleware {
	return factory(p)
}

func (p *RootProvider) Middleware(f func(*RequestProvider) httproute.Middleware) httproute.Middleware {
	return httproute.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestProvider := getRequestProvider(r)
			m := f(requestProvider)
			h := m.Handle(next)
			h.ServeHTTP(w, r)
		})
	})
}

func (p *RootProvider) Handler(f func(*RequestProvider) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestProvider := getRequestProvider(r)
		h := f(requestProvider)
		h.ServeHTTP(w, r)
	})
}

type AppProvider struct {
	*RootProvider
	Context       context.Context
	Config        *config.Config
	LoggerFactory *log.Factory
}

func (p *AppProvider) NewRequestProvider(r *http.Request) *RequestProvider {
	return &RequestProvider{
		AppProvider: p,
		Request:     r,
	}
}

type RequestProvider struct {
	*AppProvider
	Request *http.Request
}
