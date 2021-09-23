package user

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type VerificationService interface {
	IsUserVerified(identities []*identity.Info) (bool, error)
}

type Queries struct {
	*RawQueries
	Store          store
	Identities     IdentityService
	Authenticators AuthenticatorService
	Verification   VerificationService
}

func (p *Queries) Get(id string) (*model.User, error) {
	user, err := p.Store.Get(id)
	if err != nil {
		return nil, err
	}

	identities, err := p.Identities.ListByUser(id)
	if err != nil {
		return nil, err
	}

	authenticators, err := p.Authenticators.List(id)
	if err != nil {
		return nil, err
	}

	isVerified, err := p.Verification.IsUserVerified(identities)
	if err != nil {
		return nil, err
	}

	return newUserModel(user, identities, authenticators, isVerified), nil
}
