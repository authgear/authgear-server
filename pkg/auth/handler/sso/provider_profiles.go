package sso

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

// AttachProviderProfilesHandler attaches ProviderProfilesHandler to server
func AttachProviderProfilesHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/provider_profiles", &ProviderProfilesHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ProviderProfilesHandlerFactory creates ProviderProfilesHandler
type ProviderProfilesHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ProviderProfilesHandler
func (f ProviderProfilesHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ProviderProfilesHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ProviderProfilesHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type ProviderProfilesHandler struct {
	TxContext db.TxContext `dependency:"TxContext"`
}

func (h ProviderProfilesHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h ProviderProfilesHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return handler.EmptyRequestPayload{}, nil
}

// Handle function handle get oauth provider profiles request
func (h ProviderProfilesHandler) Handle(req interface{}) (resp interface{}, err error) {
	return
}
