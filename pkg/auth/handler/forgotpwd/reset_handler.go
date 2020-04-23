package forgotpwd

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type ResetPasswordProvider interface {
	ResetPassword(code string, newPassword string) error
}

// AttachForgotPasswordResetHandler attaches ForgotPasswordResetHandler to server
func AttachForgotPasswordResetHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/forgot_password/reset_password").
		Handler(auth.MakeHandler(authDependency, newResetPasswordHandler)).
		Methods("OPTIONS", "POST")
}

type ForgotPasswordResetPayload struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// nolint: gosec
// @JSONSchema
const ForgotPasswordResetRequestSchema = `
{
	"$id": "#ForgotPasswordResetRequest",
	"type": "object",
	"properties": {
		"code": { "type": "string", "minLength": 1 },
		"new_password": { "type": "string", "minLength": 1 }
	},
	"required": ["code", "new_password"]
}
`

/*
	@Operation POST /forgot_password/reset_password - Reset password
		Reset password using received recovery code.

		@Tag Forgot Password

		@RequestBody
			@JSONSchema {ForgotPasswordResetRequest}

		@Response 200 {EmptyResponse}

		@Callback password_update {PasswordUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type ResetPasswordHandler struct {
	RequireAuthz          handler.RequireAuthz
	Validator             *validation.Validator
	TxContext             db.TxContext
	ResetPasswordProvider ResetPasswordProvider
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h *ResetPasswordHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ResetPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ForgotPasswordResetPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#ForgotPasswordResetRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() (err error) {
		err = h.ResetPasswordProvider.ResetPassword(payload.Code, payload.NewPassword)
		return
	})

	resp = struct{}{}
	return
}
