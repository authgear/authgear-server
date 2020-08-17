package user

import (
	"github.com/authgear/authgear-server/pkg/lib/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type VerificationService interface {
	IsUserVerified(identities []*identity.Info, userID string) (bool, error)
	IsVerified(identities []*identity.Info, authenticators []*authenticator.Info) bool
}

type Queries struct {
	Store        store
	Identities   IdentityService
	Verification VerificationService
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

	isVerified, err := p.Verification.IsUserVerified(identities, id)
	if err != nil {
		return nil, err
	}

	return newUserModel(user, identities, isVerified), nil
}
