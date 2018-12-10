package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

func AttachChangePasswordHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/change_password", &ChangePasswordHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ChangePasswordHandlerFactory creates ChangePasswordHandler
type ChangePasswordHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new handler
func (f ChangePasswordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ChangePasswordHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ChangePasswordHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type ChangePasswordRequestPayload struct {
	Password    string `json:"password"`
	OldPassword string `json:"old_password"`
}

func (p ChangePasswordRequestPayload) Validate() error {
	return nil
}

// ChangePasswordHandler handles change password request
type ChangePasswordHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h ChangePasswordHandler) WithTx() bool {
	return true
}

// DecodeRequest decode the request payload
func (h ChangePasswordHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ChangePasswordRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

// Handle function handles the request
func (h ChangePasswordHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
