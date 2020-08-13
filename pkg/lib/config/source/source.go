package source

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Source interface {
	Open() error
	Close() error
	ProvideConfig(ctx context.Context, r *http.Request) (*config.Config, error)
}

func NewSource(
	cfg *config.ServerConfig,
	lf *LocalFile,
) Source {
	switch cfg.ConfigSource.Type {
	case config.SourceTypeLocalFile:
		return lf
	default:
		panic("config_source: invalid config source type")
	}
}
