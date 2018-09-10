package handler

import (
	"context"
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
	return handler.APIHandlerToHandler(h)
}

// MeHandler handles me request
type MeHandler struct{}

func (h MeHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.NewAllOfPolicy(
		authz.PolicyFunc(authz.DenyNoAccessKey),
		authz.PolicyFunc(authz.RequireAuthenticated),
		authz.PolicyFunc(authz.DenyDisabledUser),
	)
}

func (h MeHandler) DecodeRequest(request *http.Request) (payload handler.RequestPayload, err error) {
	payload = handler.EmptyRequestPayload{}
	return
}

func (h MeHandler) Handle(req interface{}, authInfo auth.AuthInfo) (resp interface{}, err error) {
	resp = authInfo.AuthInfo
	return
}
