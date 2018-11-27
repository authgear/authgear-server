package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachRoleRevokeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/role/revoke", &RoleRevokeHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type RoleRevokeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f RoleRevokeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &RoleRevokeHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f RoleRevokeHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	// TODO: Add OR clause to allow  master key.
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.RequireAdminRole),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type RoleRevokeRequestPayload struct {
	Roles   []string `json:"roles"`
	UserIDs []string `json:"users"`
}

func (p RoleRevokeRequestPayload) Validate() error {
	if p.Roles == nil || len(p.Roles) == 0 {
		return skyerr.NewInvalidArgument("unspecified roles in request", []string{"roles"})
	}
	if p.UserIDs == nil || len(p.UserIDs) == 0 {
		return skyerr.NewInvalidArgument("unspecified users in request", []string{"users"})
	}

	return nil
}

// RoleRevokeHandler allow system administrator to batch revoke roles from
// users
//
// RoleRevokeHandler required user with admin role.
// All specified users will have all specified roles revoked. Roles or users
// not already exisited in DB will be ignored.
//
// curl -X POST -H "Content-Type: application/json" \
//   -H "X-Skygear-API-Key: api_key" \
//   -d @- http://localhost:3000/role/revoke <<EOF
// {
//     "roles": [
//        "writer",
//        "user"
//     ],
//     "users": [
//        "95db1e34-0cc0-47b0-8a97-3948633ce09f",
//        "3df4b52b-bd58-4fa2-8aee-3d44fd7f974d"
//     ]
// }
// EOF
//
// {
//     "result": "OK"
// }
type RoleRevokeHandler struct {
	AuthContext   coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	AuthInfoStore authinfo.Store         `dependency:"AuthInfoStore"`
	AuditTrail    audit.Trail            `dependency:"AuditTrail"`
	TxContext     db.TxContext           `dependency:"TxContext"`
}

func (h RoleRevokeHandler) WithTx() bool {
	return true
}

func (h RoleRevokeHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := RoleRevokeRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h RoleRevokeHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(RoleRevokeRequestPayload)
	authInfo := h.AuthContext.AuthInfo()

	defer func() {
		if err == nil {
			h.AuditTrail.Log(audit.Entry{
				AuthID: authInfo.ID,
				Event:  audit.EventChangeRoles,
				Data: map[string]interface{}{
					"type":     "revoke",
					"user_ids": payload.UserIDs,
					"roles":    payload.Roles,
				},
			})
		}
	}()

	if err = h.AuthInfoStore.RevokeRoles(payload.UserIDs, payload.Roles); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	resp = "OK"
	return
}
