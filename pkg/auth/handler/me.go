package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/provider"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachMeHandler(
	server *server.Server,
	authDependency provider.AuthProviders,
) *server.Server {
	server.Handle("/me", &MeHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type MeHandlerFactory struct {
	Dependency provider.AuthProviders
}

func (f MeHandlerFactory) NewHandler(ctx context.Context, tenantConfig config.TenantConfiguration) handler.Handler {
	h := &MeHandler{}
	inject.DefaultInject(h, f.Dependency, ctx, tenantConfig)
	return h
}

// MeHandler handles me request
type MeHandler struct {
	auth.TokenStore    `dependency:"TokenStore"`
	auth.AuthInfoStore `dependency:"AuthInfoStore"`
}

func (h MeHandler) ProvideAuthzPolicy(r *http.Request) authz.Policy {
	return authz.NewAllOfPolicy(
		authz.PolicyFunc(authz.RequireAPIKey),
		authz.PolicyFunc(authz.RequireAuthenticated),
		authz.PolicyFunc(authz.DenyDisabledUser),
	)
}

func (h MeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx handler.AuthenticationContext) {
	output, err := json.Marshal(ctx.AuthInfo)
	if err != nil {
		// TODO:
		// handle error properly
		panic(err)
	}

	fmt.Fprint(rw, string(output))
}
