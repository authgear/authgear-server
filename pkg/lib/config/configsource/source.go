package configsource

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AppIDResolver interface {
	ResolveAppID(r *http.Request) (appID string, err error)
}

type ContextResolver interface {
	ResolveContext(appID string) (*config.AppContext, error)
}

type Handle interface {
	Open() error
	Close() error
	ReloadApp(appID string)
}

type ConfigSource struct {
	AppIDResolver   AppIDResolver
	ContextResolver ContextResolver
}

func (s *ConfigSource) ProvideContext(r *http.Request) (*config.AppContext, error) {
	appID, err := s.AppIDResolver.ResolveAppID(r)
	if err != nil {
		return nil, err
	}
	return s.ContextResolver.ResolveContext(appID)
}

type Controller struct {
	Handle          Handle
	AppIDResolver   AppIDResolver
	ContextResolver ContextResolver
}

func NewController(
	cfg *Config,
	lf *LocalFS,
	k8s *Kubernetes,
	d *Database,
) *Controller {
	switch cfg.Type {
	case TypeLocalFS:
		return &Controller{
			Handle:          lf,
			AppIDResolver:   lf,
			ContextResolver: lf,
		}
	case TypeKubernetes:
		return &Controller{
			Handle:          k8s,
			AppIDResolver:   k8s,
			ContextResolver: k8s,
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

func (c *Controller) Open() error {
	return c.Handle.Open()
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
