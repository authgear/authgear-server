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
