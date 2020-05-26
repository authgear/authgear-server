package user

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type Queries struct {
	AuthInfos    authinfo.Store
	UserProfiles userprofile.Store
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

	return newUser(p.Time.NowUTC(), &authInfo, &userProfile), nil
}
