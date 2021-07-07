package user

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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

func (p *Queries) GetRaw(id string) (*User, error) {
	return p.Store.Get(id)
}

func (p *Queries) GetManyRaw(ids []string) ([]*User, error) {
	return p.Store.GetByIDs(ids)
}

func (p *Queries) Count() (uint64, error) {
	return p.Store.Count()
}

func (p *Queries) QueryPage(sortOption SortOption, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error) {
	users, offset, err := p.Store.QueryPage(sortOption, pageArgs)
	if err != nil {
		return nil, err
	}

	var models = make([]model.PageItemRef, len(users))
	for i, u := range users {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}

		models[i] = model.PageItemRef{ID: u.ID, Cursor: cursor}
	}
	return models, nil
}
