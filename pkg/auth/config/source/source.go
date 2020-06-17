package source

import (
	"context"
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"net/http"
)

type Source interface {
	Start() error
	Shutdown() error
	ProvideConfig(ctx context.Context, r *http.Request) (*config.Config, error)
}

func NewSource(cfg *config.ServerConfig) Source {
	switch cfg.ConfigSource.Type {
	case config.SourceTypeLocalFile:
		return NewLocalFile(cfg)
	default:
		panic("config_source: invalid config source type")
	}
}
