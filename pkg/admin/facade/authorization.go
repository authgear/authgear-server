package facade

import (
	"github.com/authgear/authgear-server/pkg/lib/oauth"
)

type AuthorizationService interface {
	GetByID(id string) (*oauth.Authorization, error)
	ListByUser(userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error)
	Delete(a *oauth.Authorization) error
}

type AuthorizationFacade struct {
	Authorizations AuthorizationService
}

func (f *AuthorizationFacade) Get(id string) (*oauth.Authorization, error) {
	return f.Authorizations.GetByID(id)
}

func (f *AuthorizationFacade) List(userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error) {
	return f.Authorizations.ListByUser(userID, filters...)
}

func (f *AuthorizationFacade) Delete(a *oauth.Authorization) error {
	return f.Authorizations.Delete(a)
}
