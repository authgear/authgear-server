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

type ForgotPasswordProvider interface {
	SendCode(loginID string) error
}

// AttachForgotPasswordHandler attaches ForgotPasswordHandler to server
func AttachForgotPasswordHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/forgot_password").
		Handler(auth.MakeHandler(authDependency, newForgotPasswordHandler)).
		Methods("OPTIONS", "POST")
}

type ForgotPasswordPayload struct {
	Email string `json:"email"`
}

// nolint: gosec
// @JSONSchema
const ForgotPasswordRequestSchema = `
{
	"$id": "#ForgotPasswordRequest",
	"type": "object",
	"properties": {
		"email": { "type": "string", "format": "email" }
	},
	"required": ["email"]
}
`

/*
	@Operation POST /forgot_password - Request password recovery
		Request password recovery message to be sent to email.

		@Tag Forgot Password

		@RequestBody
			@JSONSchema {ForgotPasswordRequest}

		@Response 200 {EmptyResponse}
*/
type ForgotPasswordHandler struct {
	Validator              *validation.Validator
	TxContext              db.TxContext
	RequireAuthz           handler.RequireAuthz
	ForgotPasswordProvider ForgotPasswordProvider
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h *ForgotPasswordHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h *ForgotPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ForgotPasswordPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#ForgotPasswordRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() (err error) {
		err = h.ForgotPasswordProvider.SendCode(payload.Email)
		return
	})

	resp = struct{}{}
	return
}
