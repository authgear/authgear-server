package ssohandler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
)

// AttachCustomTokenLoginHandler attaches CustomTokenLoginHandler to server
func AttachCustomTokenLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/custom_token/login", &CustomTokenLoginHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// CustomTokenLoginHandlerFactory creates CustomTokenLoginHandler
type CustomTokenLoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new CustomTokenLoginHandler
func (f CustomTokenLoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CustomTokenLoginHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f CustomTokenLoginHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type customTokenLoginPayload struct {
}

func (payload customTokenLoginPayload) Validate() error {
	return nil
}

// CustomTokenLoginHandler handles custom login request
type CustomTokenLoginHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h CustomTokenLoginHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h CustomTokenLoginHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := customTokenLoginPayload{}
	return payload, nil
}

// Handle function handle custom token login
func (h CustomTokenLoginHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
