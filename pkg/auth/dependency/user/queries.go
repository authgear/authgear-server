package user

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type IdentityProvider interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type Queries struct {
	AuthInfos    authinfo.Store
	UserProfiles userprofile.Store
	Identities   IdentityProvider
	Time         time.Provider
}

func (p *Queries) Get(id string) (*model.User, error) {
	authInfo := authinfo.AuthInfo{}
	err := p.AuthInfos.GetAuth(id, &authInfo)
	if err != nil {
		return nil, err
	}

	userProfile, err := p.UserProfiles.GetUserProfile(id)
	if err != nil {
		return nil, err
	}

	identities, err := p.Identities.ListByUser(id)
	if err != nil {
		return nil, err
	}

	return newUser(p.Time.NowUTC(), &authInfo, &userProfile, identities), nil
}
