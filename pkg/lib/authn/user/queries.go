package user

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type VerificationService interface {
	IsUserVerified(identities []*identity.Info) (bool, error)
	DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
}

type Queries struct {
	*RawQueries
	Store          store
	Identities     IdentityService
	Authenticators AuthenticatorService
	Verification   VerificationService
}

func (p *Queries) Get(id string, role accesscontrol.Role) (*model.User, error) {
	user, err := p.RawQueries.GetRaw(id)
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

	stdAttrs, err := p.Verification.DeriveStandardAttributes(role, id, user.UpdatedAt, user.StandardAttributes)
	if err != nil {
		return nil, err
	}

	return newUserModel(user, identities, authenticators, isVerified, stdAttrs), nil
}
