package oauth

import (
	"context"
)

type AuthorizationStore interface {
	Get(ctx context.Context, userID, clientID string) (*Authorization, error)
	GetByID(ctx context.Context, id string) (*Authorization, error)
	ListByUserID(ctx context.Context, userID string) ([]*Authorization, error)
	Create(ctx context.Context, a *Authorization) error
	Delete(ctx context.Context, a *Authorization) error
	ResetAll(ctx context.Context, userID string) error
	UpdateScopes(ctx context.Context, a *Authorization) error
}
