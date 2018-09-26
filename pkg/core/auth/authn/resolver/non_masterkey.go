package resolver

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type nonMasterkeyAuthContextResolver struct {
	TokenStore    authtoken.Store
	AuthInfoStore authinfo.Store
}

func (r nonMasterkeyAuthContextResolver) Resolve(req *http.Request) (ctx handler.AuthContext, err error) {
	tokenStr := model.GetAccessToken(req)

	token := &authtoken.Token{}
	err = r.TokenStore.Get(tokenStr, token)
	if err != nil {
		// TODO:
		// handle error properly
		return
	}

	ctx.Token = token

	info := &authinfo.AuthInfo{}
	err = r.AuthInfoStore.GetAuth(token.AuthInfoID, info)
	if err != nil {
		// TODO:
		// handle error properly
		return
	}

	ctx.AuthInfo = info

	return
}
