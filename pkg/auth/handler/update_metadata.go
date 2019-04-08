package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachUpdateMetadataHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/update_metadata", &UpdateMetadataHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type UpdateMetadataHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f UpdateMetadataHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &UpdateMetadataHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

func (f UpdateMetadataHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type UpdateMetadataRequestPayload struct {
	UserID   string                 `json:"user_id"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (p UpdateMetadataRequestPayload) Validate() error {
	if p.UserID == "" {
		return skyerr.NewInvalidArgument("empty user_id", []string{"user_id"})
	}

	return nil
}

// UpdateMetadataHandler handles update current user's metadata
//
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/update_metadata <<EOF
//  {
//    "metadata": {
//    }
//  }
//  EOF
//
// {
//   "user_id": "3df4b52b-bd58-4fa2-8aee-3d44fd7f974d",
//   "last_login_at": "2016-09-08T06:42:59.871181Z",
//   "last_seen_at": "2016-09-08T07:15:18.026567355Z",
//   "metadata": {}
// }
type UpdateMetadataHandler struct {
	AuthContext          coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	AuthInfoStore        authinfo.Store         `dependency:"AuthInfoStore"`
	TxContext            db.TxContext           `dependency:"TxContext"`
	UserProfileStore     userprofile.Store      `dependency:"UserProfileStore"`
	PasswordAuthProvider password.Provider      `dependency:"PasswordAuthProvider"`
}

func (h UpdateMetadataHandler) WithTx() bool {
	return true
}

func (h UpdateMetadataHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := UpdateMetadataRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h UpdateMetadataHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(UpdateMetadataRequestPayload)
	keyType := h.AuthContext.AccessKeyType()
	currentUserID := h.AuthContext.AuthInfo().ID
	updateUserID := payload.UserID
	inMetadata := payload.Metadata

	if keyType == model.APIAccessKey && currentUserID != payload.UserID {
		err = skyerr.NewError(skyerr.PermissionDenied, "Unable to update another user's metadata")
		return
	}

	authInfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(updateUserID, &authInfo); e != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
		// TODO: more error handling here if necessary
		err = skyerr.NewResourceFetchFailureErr("auth_data", payload.UserID)
		return
	}

	var profile userprofile.UserProfile
	if profile, err = h.UserProfileStore.UpdateUserProfile(authInfo.ID, &authInfo, inMetadata); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to update user profile")
		return
	}

	token := h.AuthContext.Token().AccessToken

	respFactory := response.AuthResponseFactory{
		PasswordAuthProvider: h.PasswordAuthProvider,
	}
	resp = respFactory.NewAuthResponse(authInfo, profile, token)

	return
}
