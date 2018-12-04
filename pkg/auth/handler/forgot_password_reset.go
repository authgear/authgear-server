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
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AttachForgotPasswordResetHandler attaches ForgotPasswordResetHandler to server
func AttachForgotPasswordResetHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/forgot_password/reset_password", &ForgotPasswordResetHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ForgotPasswordResetHandlerFactory creates ForgotPasswordResetHandler
type ForgotPasswordResetHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordResetHandler
func (f ForgotPasswordResetHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordResetHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ForgotPasswordResetHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type ForgotPasswordResetPayload struct {
}

func (payload ForgotPasswordResetPayload) Validate() error {
	return nil
}

// ForgotPasswordResetHandler reset user password with given code from email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/forgot_password/reset_password <<EOF
//  {
//    "user_id": "xxx",
//    "code": "xxx",
//    "expire_at": xxx, (utc timestamp)
//    "new_password": "xxx",
//  }
//  EOF
type ForgotPasswordResetHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h ForgotPasswordResetHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h ForgotPasswordResetHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ForgotPasswordResetPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

func (h ForgotPasswordResetHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
