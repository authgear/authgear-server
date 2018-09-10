package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/provider"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachSignupHandler(
	server *server.Server,
	authDependency provider.AuthProviders,
) *server.Server {
	server.Handle("/signup", &SignupHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type SignupHandlerFactory struct {
	Dependency provider.AuthProviders
}

func (f SignupHandlerFactory) NewHandler(ctx context.Context, tenantConfig config.TenantConfiguration) handler.Handler {
	h := &SignupHandler{}
	inject.DefaultInject(h, f.Dependency, ctx, tenantConfig)
	return handler.APIHandlerToHandler(h)
}

type SignupRequestPayload struct {
	AuthDataData     map[string]interface{} `json:"auth_data"`
	Password         string                 `json:"password"`
	Provider         string                 `json:"provider"`
	ProviderAuthData map[string]interface{} `json:"provider_auth_data"`
	RawProfile       map[string]interface{} `json:"profile"`
}

func (p SignupRequestPayload) Validate() error {
	return nil
}

// SignupHandler handles signup request
type SignupHandler struct{}

func (h SignupHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(authz.DenyNoAccessKey)
}

func (h SignupHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SignupRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h SignupHandler) Handle(req interface{}, authInfo auth.AuthInfo) (resp interface{}, err error) {
	payload := req.(SignupRequestPayload)
	// TODO:
	resp = payload
	return
}
