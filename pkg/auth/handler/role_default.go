package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachRoleDefaultHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/role/default", &RoleDefaultHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RoleDefaultHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RoleDefaultHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RoleDefaultHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f RoleDefaultHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		policy.AnyOf(
			authz.PolicyFunc(policy.RequireAdminRole),
			authz.PolicyFunc(policy.RequireMasterKey),
		),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type RoleDefaultRequestPayload struct {
	Roles []string `json:"roles"`
}

func (p RoleDefaultRequestPayload) Validate() error {
	if p.Roles == nil || len(p.Roles) == 0 {
		return skyerr.NewInvalidArgument("unspecified roles in request", []string{"roles"})
	}

	return nil
}

// RoleDefaultHandler enable system administrator to set default user role on
// signup
//
// curl -X POST -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: MASTER_KEY" \
//   -d @- http://localhost:3000/role/default <<EOF
// {
//     "roles": [
//        "writer",
//        "user"
//     ]
// }
// EOF
//
// {
//     "result": [
//        "writer",
//        "user"
//     ]
// }
type RoleDefaultHandler struct {
	RoleStore role.Store    `dependency:"RoleStore"`
	Logger    *logrus.Entry `dependency:"HandlerLogger"`
	TxContext db.TxContext  `dependency:"TxContext"`
}

func (h RoleDefaultHandler) WithTx() bool {
	return true
}

func (h RoleDefaultHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := RoleDefaultRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h RoleDefaultHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(RoleDefaultRequestPayload)

	if _, err = role.EnsureRole(h.RoleStore, h.Logger, payload.Roles); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	if err = h.RoleStore.SetDefaultRoles(payload.Roles); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	resp = payload.Roles
	return
}
