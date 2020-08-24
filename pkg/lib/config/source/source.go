package source

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type ServerType string

const (
	ServerTypeMain     ServerType = "main"
	ServerTypeResolver ServerType = "resolver"
	ServerTypeAdminAPI ServerType = "admin_api"
)

type source interface {
	Open() error
	Close() error
	ProvideConfig(ctx context.Context, r *http.Request, server ServerType) (*config.Config, error)
}

type ConfigSource struct {
	src        source
	serverType ServerType
}

func (s *ConfigSource) ProvideConfig(ctx context.Context, r *http.Request) (*config.Config, error) {
	return s.src.ProvideConfig(ctx, r, s.serverType)
}

type Controller struct {
	src source
}

func NewController(
	cfg *config.ServerConfig,
	lf *LocalFS,
) *Controller {
	switch cfg.ConfigSource.Type {
	case config.SourceTypeLocalFS:
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

func (c *Controller) ForServer(server ServerType) *ConfigSource {
	return &ConfigSource{src: c.src, serverType: server}
}
