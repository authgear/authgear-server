package configsource

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AppIDResolver interface {
	ResolveAppID(ctx context.Context, r *http.Request) (appID string, err error)
}

type ContextResolver interface {
	ResolveContext(ctx context.Context, appID string) (*config.AppContext, error)
}

type Handle interface {
	Open(ctx context.Context) error
	Close() error
	ReloadApp(appID string)
}

type ConfigSource struct {
	AppIDResolver   AppIDResolver
	ContextResolver ContextResolver
}

func (s *ConfigSource) ProvideContext(ctx context.Context, r *http.Request) (*config.AppContext, error) {
	appID, err := s.AppIDResolver.ResolveAppID(ctx, r)
	if err != nil {
		return nil, err
	}
	return s.ContextResolver.ResolveContext(ctx, appID)
}

type Controller struct {
	Handle          Handle
	AppIDResolver   AppIDResolver
	ContextResolver ContextResolver
}

func NewController(
	cfg *Config,
	lf *LocalFS,
	d *Database,
) *Controller {
	switch cfg.Type {
	case TypeLocalFS:
		return &Controller{
			Handle:          lf,
			AppIDResolver:   lf,
			ContextResolver: lf,
		}
	case TypeDatabase:
		return &Controller{
			Handle:          d,
			AppIDResolver:   d,
			ContextResolver: d,
		}
	default:
		panic("config_source: invalid config source type")
	}
}

func (c *Controller) Open(ctx context.Context) error {
	return c.Handle.Open(ctx)
}

func (c *Controller) Close() error {
	return c.Handle.Close()
}

func (c *Controller) ReloadApp(appID string) {
	c.Handle.ReloadApp(appID)
}

func (c *Controller) GetConfigSource() *ConfigSource {
	return &ConfigSource{
		AppIDResolver:   c.AppIDResolver,
		ContextResolver: c.ContextResolver,
	}
}

// ResolveContext allows direct resolution from appID.
// It is useful when you get appID somewhere else, rather than from a HTTP request.
func (c *Controller) ResolveContext(ctx context.Context, appID string) (*config.AppContext, error) {
	return c.ContextResolver.ResolveContext(ctx, appID)
}
