package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	authModel "github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachUpdateMetadataHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/update_metadata").
		Handler(server.FactoryToHandler(&UpdateMetadataHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type UpdateMetadataHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f UpdateMetadataHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &UpdateMetadataHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type UpdateMetadataRequestPayload struct {
	UserID      string                 `json:"user_id"`
	Metadata    map[string]interface{} `json:"metadata"`
	isMasterKey bool
}

// @JSONSchema
const UpdateMetadataRequestSchema = `
{
	"$id": "#UpdateMetadataRequest",
	"type": "object",
	"properties": {
		"user_id": { "type": "string" },
		"metadata": { "type": "object" }
	},
	"required": ["metadata"]
}
`

func (p *UpdateMetadataRequestPayload) Validate() []validation.ErrorCause {
	if p.isMasterKey {
		if p.UserID == "" {
			return []validation.ErrorCause{{
				Kind:    validation.ErrorRequired,
				Pointer: "/user_id",
				Message: "user_id is required",
			}}
		}
	}
	return nil
}

/*
	@Operation POST /update_metadata - Update metadata
		Changes metadata of current user.
		If master key is used as access key, other users can be specified.

		@Tag User
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			Describe target user and new metadata.
			@JSONSchema {UpdateMetadataRequest}

		@Response 200
			User information with new metadata.
			@JSONSchema {UserResponse}

		@Callback user_update {UserUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type UpdateMetadataHandler struct {
	Validator        *validation.Validator `dependency:"Validator"`
	RequireAuthz     handler.RequireAuthz  `dependency:"RequireAuthz"`
	AuthInfoStore    authinfo.Store        `dependency:"AuthInfoStore"`
	TxContext        db.TxContext          `dependency:"TxContext"`
	UserProfileStore userprofile.Store     `dependency:"UserProfileStore"`
	HookProvider     hook.Provider         `dependency:"HookProvider"`
}

func (h UpdateMetadataHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUserOrMasterKey
}

func (h UpdateMetadataHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	result, err := h.Handle(resp, req)
	if err == nil {
		handler.WriteResponse(resp, handler.APIResponse{Result: result})
	} else {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
	}
}

func (h UpdateMetadataHandler) Handle(resp http.ResponseWriter, req *http.Request) (result interface{}, err error) {
	var payload UpdateMetadataRequestPayload
	if err = handler.BindJSONBody(req, resp, h.Validator, "#UpdateMetadataRequest", &payload); err != nil {
		return
	}

	err = db.WithTx(h.TxContext, func() error {
		accessKey := coreAuth.GetAccessKey(req.Context())

		var targetUserID string
		if accessKey.IsMasterKey {
			targetUserID = payload.UserID
		} else {
			if payload.UserID != "" {
				err = skyerr.NewForbidden("must not specify user_id")
				return err
			}
			targetUserID = auth.GetSession(req.Context()).AuthnAttrs().UserID
		}

		newMetadata := payload.Metadata

		authInfo := authinfo.AuthInfo{}
		if err = h.AuthInfoStore.GetAuth(targetUserID, &authInfo); err != nil {
			return err
		}

		var oldProfile, newProfile userprofile.UserProfile
		if oldProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
			return err
		}

		if newProfile, err = h.UserProfileStore.UpdateUserProfile(authInfo.ID, newMetadata); err != nil {
			return err
		}

		oldUser := authModel.NewUser(authInfo, oldProfile)
		user := authModel.NewUser(authInfo, newProfile)

		err = h.HookProvider.DispatchEvent(
			event.UserUpdateEvent{
				Reason:   event.UserUpdateReasonUpdateMetadata,
				User:     oldUser,
				Metadata: &newProfile.Data,
			},
			&user,
		)
		if err != nil {
			return err
		}

		result = authModel.NewAuthResponse(user)
		return nil
	})
	return
}
