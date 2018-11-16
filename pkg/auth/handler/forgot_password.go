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

// AttachForgotPasswordHandler attaches ForgotPasswordHandler to server
func AttachForgotPasswordHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/forgot_password", &ForgotPasswordHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	server.Handle("/forgot_password/test", &ForgotPasswordTestHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ForgotPasswordHandlerFactory creates ForgotPasswordHandler
type ForgotPasswordHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordHandler
func (f ForgotPasswordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ForgotPasswordHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type ForgotPasswordPayload struct {
}

func (payload ForgotPasswordPayload) Validate() error {
	return nil
}

// ForgotPasswordHandler send a reset password email to given email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/forgot_password <<EOF
//  {
//     "email": "xxx@oursky.com"
//  }
//  EOF
type ForgotPasswordHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h ForgotPasswordHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h ForgotPasswordHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ForgotPasswordPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

// Handle function handle set disabled request
func (h ForgotPasswordHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}

// ForgotPasswordTestHandlerFactory creates ForgotPasswordTestHandler
type ForgotPasswordTestHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordTestHandler
func (f ForgotPasswordTestHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordTestHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, nil)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ForgotPasswordTestHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

type ForgotPasswordTestPayload struct {
}

func (payload ForgotPasswordTestPayload) Validate() error {
	return nil
}

// ForgotPasswordTestHandler send a dummy reset password email to given email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/forgot_password/test <<EOF
//  {
//     "email": "xxx@oursky.com",
//     "text_template": "xxx",
//     "html_template": "xxx",
//     "subject": "xxx",
//     "sender": "xxx",
//     "reply_to": "xxx",
//     "sender_name": "xxx",
//     "reply_to_name": "xxx"
//  }
//  EOF
type ForgotPasswordTestHandler struct {
}

func (h ForgotPasswordTestHandler) WithTx() bool {
	return false
}

// DecodeRequest decode request payload
func (h ForgotPasswordTestHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ForgotPasswordTestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

// Handle function handle set disabled request
func (h ForgotPasswordTestHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
