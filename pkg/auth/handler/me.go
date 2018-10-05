package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/handler/context"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachMeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/me", &MeHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type MeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f MeHandlerFactory) NewHandler(request *http.Request) handler.Handler {
	h := &MeHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h)
}

func (f MeHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

// MeHandler handles me request
type MeHandler struct{}

func (h MeHandler) DecodeRequest(request *http.Request) (payload handler.RequestPayload, err error) {
	payload = handler.EmptyRequestPayload{}
	return
}

func (h MeHandler) Handle(req interface{}, ctx context.AuthContext) (resp interface{}, err error) {
	resp = ctx.AuthInfo
	return
}
