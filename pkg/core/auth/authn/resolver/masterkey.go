package resolver

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type masterkeyAuthContextResolver struct {
	TokenStore    authtoken.Store
	AuthInfoStore authinfo.Store
}

func (r masterkeyAuthContextResolver) Resolve(req *http.Request) (token *authtoken.Token, authInfo *authinfo.AuthInfo, err error) {
	tokenStr := model.GetAccessToken(req)
	token = &authtoken.Token{}
	r.TokenStore.Get(tokenStr, token)

	if token.AuthInfoID == "" {
		token.AuthInfoID = "_god"
	}

	info := &authinfo.AuthInfo{}
	if err = r.AuthInfoStore.GetAuth(token.AuthInfoID, info); err == skydb.ErrUserNotFound {
		info.ID = token.AuthInfoID

		if err = r.AuthInfoStore.CreateAuth(info); err == skydb.ErrUserDuplicated {
			// user already exists, error can be ignored
			err = nil
		}
	}

	authInfo = info

	return
}
