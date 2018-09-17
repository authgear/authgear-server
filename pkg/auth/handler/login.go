package handler

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/login", &LoginHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type LoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f LoginHandlerFactory) NewHandler(ctx context.Context, tenantConfig config.TenantConfiguration) handler.Handler {
	h := &LoginHandler{}
	inject.DefaultInject(h, f.Dependency, ctx, tenantConfig)
	return handler.APIHandlerToHandler(h)
}

// LoginHandler handles login request
type LoginHandler struct{}

func (h LoginHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

func (h LoginHandler) DecodeRequest(request *http.Request) (payload handler.RequestPayload, err error) {
	payload = handler.EmptyRequestPayload{}
	return
}

func (h LoginHandler) Handle(req interface{}, ctx handler.AuthContext) (resp interface{}, err error) {
	resp = map[string]interface{}{"user": "user:abc"}
	return
}
