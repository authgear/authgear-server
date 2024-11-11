package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuthenticatorLoaderAuthenticatorService interface {
	GetMany(ctx context.Context, ids []string) ([]*authenticator.Info, error)
}

type AuthenticatorLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Authenticators AuthenticatorLoaderAuthenticatorService
}

func NewAuthenticatorLoader(authenticators AuthenticatorLoaderAuthenticatorService) *AuthenticatorLoader {
	l := &AuthenticatorLoader{
		Authenticators: authenticators,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *AuthenticatorLoader) LoadFunc(ctx context.Context, keys []interface{}) ([]interface{}, error) {
	// Prepare IDs.
	ids := make([]string, len(keys))
	for i, key := range keys {
		ids[i] = key.(string)
	}

	// Get entities.
	entities, err := l.Authenticators.GetMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*authenticator.Info)
	for _, entity := range entities {
		entityMap[entity.ID] = entity
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, key := range keys {
		entity := entityMap[key.(string)]
		out[i] = entity
	}
	return out, nil
}
