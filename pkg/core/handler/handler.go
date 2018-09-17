package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type AuthContext struct {
	AccessKeyType model.KeyType
	AuthInfo      *authinfo.AuthInfo
	Token         *authtoken.Token
}

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, AuthContext)
}

type HandlerFunc func(http.ResponseWriter, *http.Request, AuthContext)

func (f HandlerFunc) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx AuthContext) {
	f(rw, r, ctx)
}
