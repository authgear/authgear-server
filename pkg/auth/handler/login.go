package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/db"
	"github.com/skygeario/skygear-server/pkg/auth/provider"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachLoginHandler(
	server *server.Server,
	authDependency provider.AuthProviders,
) *server.Server {
	server.Handle("/login", &LoginHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type LoginHandlerFactory struct {
	Dependency provider.AuthProviders
}

func (f LoginHandlerFactory) NewHandler(ctx context.Context, tenantConfig config.TenantConfiguration) handler.Handler {
	h := &LoginHandler{}
	inject.DefaultInject(h, f.Dependency, ctx, tenantConfig)
	return h
}

// LoginHandler handles login request
type LoginHandler struct {
	DB *db.DBConn `dependency:"DB"`
}

func (h LoginHandler) ProvideAuthzPolicy(r *http.Request) authz.Policy {
	return authz.PolicyFunc(authz.DenyNoAccessKey)
}

func (h LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	input, _ := ioutil.ReadAll(r.Body)
	fmt.Fprintln(rw, `{"user": "`+h.DB.GetRecord("user:"+string(input))+`"}`)
}
