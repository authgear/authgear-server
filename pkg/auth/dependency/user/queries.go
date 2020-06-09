package user

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type IdentityProvider interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type Queries struct {
	Store      store
	Identities IdentityProvider
	Time       time.Provider
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

	return newUserModel(user, identities), nil
}
