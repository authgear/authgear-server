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

type Source interface {
	Open() error
	Close() error
	ProvideConfig(ctx context.Context, r *http.Request, server ServerType) (*config.Config, error)
}

func NewSource(
	cfg *config.ServerConfig,
	lf *LocalFS,
) Source {
	switch cfg.ConfigSource.Type {
	case config.SourceTypeLocalFS:
		return lf
	default:
		panic("config_source: invalid config source type")
	}
}
