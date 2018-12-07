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

func AttachRoleAdminHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/role/admin", &RoleAdminHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RoleAdminHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RoleAdminHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RoleAdminHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f RoleAdminHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type RoleAdminRequestPayload struct {
	Roles []string `json:"roles"`
}

func (p RoleAdminRequestPayload) Validate() error {
	if p.Roles == nil || len(p.Roles) == 0 {
		return skyerr.NewInvalidArgument("unspecified roles in request", []string{"roles"})
	}

	return nil
}

// RoleAdminHandler enable system administrator to set which roles can perform
// administrative action, like change others user role.
//
// curl \
//   -X POST \
//   -H "X-Skygear-Api-Key: MASTER_KEY" \
//   -H "X-Skygear-Access-Token: ACCESS_TOKEN" \
//   -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/role/admin
// <<EOF
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
type RoleAdminHandler struct {
	RoleStore role.Store    `dependency:"RoleStore"`
	Logger    *logrus.Entry `dependency:"HandlerLogger"`
	TxContext db.TxContext  `dependency:"TxContext"`
}

func (h RoleAdminHandler) WithTx() bool {
	return true
}

func (h RoleAdminHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := RoleAdminRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h RoleAdminHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(RoleAdminRequestPayload)

	if _, err = role.EnsureRole(h.RoleStore, h.Logger, payload.Roles); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	if err = h.RoleStore.SetAdminRoles(payload.Roles); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	resp = payload.Roles
	return
}
