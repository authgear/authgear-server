package accountdeletion

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AppContextResolver interface {
	ResolveContext(appID string) (*config.AppContext, error)
}

type Runnable struct {
	Store              *Store
	AppContextResolver AppContextResolver
}

func (r *Runnable) Run(ctx context.Context) error {
	return nil
}
