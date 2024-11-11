package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type AuthorizationService interface {
	GetByID(ctx context.Context, id string) (*oauth.Authorization, error)
	ListByUser(ctx context.Context, userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error)
	Delete(ctx context.Context, a *oauth.Authorization) error
}

type AuthorizationFacade struct {
	Authorizations AuthorizationService
}

func (f *AuthorizationFacade) Get(ctx context.Context, id string) (*oauth.Authorization, error) {
	return f.Authorizations.GetByID(ctx, id)
}

func (f *AuthorizationFacade) List(ctx context.Context, userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error) {
	return f.Authorizations.ListByUser(ctx, userID, filters...)
}

func (f *AuthorizationFacade) Delete(ctx context.Context, a *oauth.Authorization) error {
	return f.Authorizations.Delete(ctx, a)
}
