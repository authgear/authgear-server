package configsource

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AppIDResolver interface {
	ResolveAppID(r *http.Request) (appID string, err error)
}

type ConfigGetter interface {
	GetConfig(appID string) (*config.Config, error)
}

type Handle interface {
	Open() error
	Close() error
}

type ConfigSource struct {
	AppIDResolver AppIDResolver
	ConfigGetter  ConfigGetter
}

func (s *ConfigSource) ProvideConfig(r *http.Request) (*config.Config, error) {
	appID, err := s.AppIDResolver.ResolveAppID(r)
	if err != nil {
		return nil, err
	}
	return s.ConfigGetter.GetConfig(appID)
}

type Controller struct {
	Handle        Handle
	AppIDResolver AppIDResolver
	ConfigGetter  ConfigGetter
}

func NewController(
	cfg *Config,
	lf *LocalFS,
) *Controller {
	switch cfg.Type {
	case TypeLocalFS:
		return &Controller{
			Handle:        lf,
			AppIDResolver: lf,
			ConfigGetter:  lf,
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

func (c *Controller) GetConfigSource() *ConfigSource {
	return &ConfigSource{
		AppIDResolver: c.AppIDResolver,
		ConfigGetter:  c.ConfigGetter,
	}
}
