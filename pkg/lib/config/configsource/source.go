package configsource

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type source interface {
	Open() error
	Close() error
	ProvideConfig(r *http.Request) (*config.Config, error)
}

type ConfigSource struct {
	src source
}

func (s *ConfigSource) ProvideConfig(r *http.Request) (*config.Config, error) {
	return s.src.ProvideConfig(r)
}

type Controller struct {
	src source
}

func NewController(
	cfg *Config,
	lf *LocalFS,
) *Controller {
	switch cfg.Type {
	case TypeLocalFS:
		return &Controller{src: lf}
	default:
		panic("config_source: invalid config source type")
	}
}

func (c *Controller) Open() error {
	return c.src.Open()
}

func (c *Controller) Close() error {
	return c.src.Close()
}

func (c *Controller) GetConfigSource() *ConfigSource {
	return &ConfigSource{src: c.src}
}
