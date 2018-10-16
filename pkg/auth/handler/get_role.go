package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachGetRoleHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/role/get", &GetRoleHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type GetRoleHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f GetRoleHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &GetRoleHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f GetRoleHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		// FIXME: this endpoint should hanlde request with master key or with access key
		// Users can only get his own roles except that administrators can query roles
		// of other users.
		// This is temporary implementation to support admin user role only.
		authz.PolicyFunc(policy.RequireMasterKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type GetRoleRequestPayload struct {
	UserIDs []string `json:"users"`
}

func (p GetRoleRequestPayload) Validate() error {
	return nil
}

// GetRoleHandler returns roles of users specified by user IDs. Users can only
// get his own roles except that administrators can query roles of other users.
//
// curl \
//   -X POST \
//   -H "Content-Type: application/json" \
//   -H "X-Skygear-Api-Key: MASTER_KEY" \
//   -H "X-Skygear-Access-Token: ACCESS_TOKEN" \
//   -d @- \
//   http://localhost:3000/role/get \
// <<EOF
// {
//     "users": [
//        "user_id_1",
//        "user_id_2",
//     ]
// }
// EOF
//
// {
//     "result": {
//         "user_id_1": [
//             "developer",
//         ],
//         "user_id_2": [
//         ],
//     }
// }
type GetRoleHandler struct {
	AuthInfoStore authinfo.Store `dependency:"AuthInfoStore"`
	TxContext     db.TxContext   `dependency:"TxContext"`
}

func (h GetRoleHandler) WithTx() bool {
	return true
}

func (h GetRoleHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := GetRoleRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

// Handle getting roles of users specified by user IDs. Users can only
// get his own roles except that administrators can query roles of other users.
// TODO: currently not able to query role of oneself.
func (h GetRoleHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(GetRoleRequestPayload)
	roleMap, err := h.AuthInfoStore.GetRoles(payload.UserIDs)
	if err != nil {
		err = skyerr.NewError(skyerr.UnexpectedError, "GetRoles failed")
		return
	}
	resp = roleMap
	return
}
