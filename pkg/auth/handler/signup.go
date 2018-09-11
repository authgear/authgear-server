package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/server/audit"

	"github.com/skygeario/skygear-server/pkg/auth/provider"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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
	AuthData         map[string]interface{} `json:"auth_data"`
	Password         string                 `json:"password"`
	Provider         string                 `json:"provider"`
	ProviderAuthData map[string]interface{} `json:"provider_auth_data"`
	// TODO:
	// RawProfile       map[string]interface{} `json:"profile"`
}

func (p SignupRequestPayload) Validate() error {
	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

// SignupHandler handles signup request
type SignupHandler struct {
	AuthDataChecker provider.AuthDataChecker `dependency:"AuthDataChecker"`
	PasswordChecker provider.PasswordChecker `dependency:"PasswordChecker"`
}

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

	if valid := h.AuthDataChecker.IsValid(payload.AuthData); !valid {
		err = skyerr.NewInvalidArgument("invalid auth data", []string{"auth_data"})
		return
	}

	// TODO: check duplicated keys in auth data and profile

	// validate password
	if err = h.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: payload.Password,
	}); err != nil {
		return
	}

	// TODO:
	resp = payload
	return
}
