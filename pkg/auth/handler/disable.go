package handler

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	authModel "github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AttachSetDisableHandler attaches SetDisableHandler to server
func AttachSetDisableHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/disable/set", &SetDisableHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// SetDisableHandlerFactory creates SetDisableHandler
type SetDisableHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new SetDisableHandler
func (f SetDisableHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SetDisableHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return h.RequireAuthz(handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext), h)
}

type setDisableUserPayload struct {
	UserID       string `json:"user_id"`
	Disabled     bool   `json:"disabled"`
	Message      string `json:"message"`
	ExpiryString string `json:"expiry"`
	expiry       *time.Time
}

// @JSONSchema
const SetDisableRequestSchema = `
{
	"$id": "#SetDisableRequest",
	"type": "object",
	"properties": {
		"auth_id": { "type": "string" },
		"disabled": { "type": "boolean" },
		"message": { "type": "string" },
		"expiry": { "type": "string" }
	}
}
`

func (payload setDisableUserPayload) Validate() error {
	if payload.UserID == "" {
		// TODO(error): JSON schema
		return skyerr.NewInvalid("invalid user ID")
	}
	return nil
}

/*
	@Operation POST /disable/set - Set user disabled status
		Disable/enable target user.

		@Tag Administration
		@SecurityRequirement master_key
		@SecurityRequirement access_token

		@RequestBody
			Describe target user and desired disable status.
			@JSONSchema {SetDisableRequest}
			@JSONExample EnableUser - Enable user
				{
					"auth_id": "F1D4AAAC-A31A-4471-92B2-6E08376BDD87",
					"disabled": false
				}
			@JSONExample DisableUser - Disable user permanently
				{
					"auth_id": "F1D4AAAC-A31A-4471-92B2-6E08376BDD87",
					"disabled": true
				}
			@JSONExample DisableUserExpiry - Disable user with expiry
				{
					"auth_id": "F1D4AAAC-A31A-4471-92B2-6E08376BDD87",
					"disabled": true,
					"message": "Banned",
					"expiry": "2019-07-31T09:39:22.349Z"
				}

		@Response 200 {EmptyResponse}

		@Callback user_update {UserUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type SetDisableHandler struct {
	AuthContext      coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	RequireAuthz     handler.RequireAuthz   `dependency:"RequireAuthz"`
	AuthInfoStore    authinfo.Store         `dependency:"AuthInfoStore"`
	UserProfileStore userprofile.Store      `dependency:"UserProfileStore"`
	AuditTrail       audit.Trail            `dependency:"AuditTrail"`
	HookProvider     hook.Provider          `dependency:"HookProvider"`
	TxContext        db.TxContext           `dependency:"TxContext"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h SetDisableHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireMasterKey),
		policy.RequireValidUser,
	)
}

func (h SetDisableHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h SetDisableHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := setDisableUserPayload{}
	if err := handler.DecodeJSONBody(request, resp, &payload); err != nil {
		return nil, err
	}

	if payload.ExpiryString != "" {
		if expiry, err := time.Parse(time.RFC3339, payload.ExpiryString); err == nil {
			payload.expiry = &expiry
		} else {
			// TODO(error): JSON schema
			return nil, skyerr.NewInvalid("invalid expiry")
		}
	}

	return payload, nil
}

// Handle function handle set disabled request
func (h SetDisableHandler) Handle(req interface{}) (resp interface{}, err error) {
	p := req.(setDisableUserPayload)

	info := authinfo.AuthInfo{}
	if err = h.AuthInfoStore.GetAuth(p.UserID, &info); err != nil {
		return
	}

	profile, err := h.UserProfileStore.GetUserProfile(info.ID)
	if err != nil {
		return
	}

	oldUser := authModel.NewUser(info, profile)

	info.Disabled = p.Disabled
	if !info.Disabled {
		info.DisabledMessage = ""
		info.DisabledExpiry = nil
	} else {
		info.DisabledMessage = p.Message
		info.DisabledExpiry = p.expiry
	}

	if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
		return
	}

	user := authModel.NewUser(info, profile)

	err = h.HookProvider.DispatchEvent(
		event.UserUpdateEvent{
			Reason:     event.UserUpdateReasonAdministrative,
			User:       oldUser,
			IsDisabled: &p.Disabled,
		},
		&user,
	)
	if err != nil {
		return
	}

	h.logAuditTrail(p)

	resp = map[string]string{}

	return
}

func (h SetDisableHandler) logAuditTrail(p setDisableUserPayload) {
	var event audit.Event
	if p.Disabled {
		event = audit.EventDisableUser
	} else {
		event = audit.EventEnableUser
	}

	h.AuditTrail.Log(audit.Entry{
		UserID: p.UserID,
		Event:  event,
	})
}
