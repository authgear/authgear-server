package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	interactionflows "github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreauth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachSignupHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/signup").
		Handler(auth.MakeHandler(authDependency, newSignupHandler)).
		Methods("OPTIONS", "POST")
}

type SignupRequestPayload struct {
	LoginIDs        []loginid.LoginID      `json:"login_ids"`
	Password        string                 `json:"password"`
	Metadata        map[string]interface{} `json:"metadata"`
	OnUserDuplicate model.OnUserDuplicate  `json:"on_user_duplicate"`
}

// @JSONSchema
const SignupRequestSchema = `
{
	"$id": "#SignupRequest",
	"type": "object",
	"properties": {
		"login_ids": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"key": { "type": "string", "minLength": 1 },
					"value": { "type": "string", "minLength": 1 }
				},
				"required": ["key", "value"]
			},
			"minItems": 1,
			"maxItems": 1
		},
		"password": { "type": "string", "minLength": 1 },
		"metadata": { "type": "object" },
		"on_user_duplicate": {
			"type": "string",
			"enum": ["abort", "create"]
		}
	},
	"required": ["login_ids", "password"]
}
`

func (p *SignupRequestPayload) SetDefaultValue() {
	if p.OnUserDuplicate == "" {
		p.OnUserDuplicate = model.OnUserDuplicateDefault
	}
	if p.Metadata == nil {
		// Avoid { metadata: null } in the response user object
		p.Metadata = make(map[string]interface{})
	}
}

type SignupInteractionFlow interface {
	SignupWithLoginIDPassword(
		clientID string,
		loginIDKey string,
		loginID string,
		password string,
		metadata map[string]interface{},
		onUserDuplicate model.OnUserDuplicate,
	) (*interactionflows.AuthResult, error)
}

/*
	@Operation POST /signup - Signup using password
		Signup user with login IDs and password.

		@Tag User

		@RequestBody
			Describe login IDs, password, and initial metadata.
			@JSONSchema {SignupRequest}

		@Response 200
			Signed up user and access token.
			@JSONSchema {AuthResponse}

		@Callback user_create {UserCreateEvent}
		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type SignupHandler struct {
	Validator    *validation.Validator
	Interactions SignupInteractionFlow
	TxContext    db.TxContext
}

func (h SignupHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
}

func (h SignupHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload SignupRequestPayload, err error) {
	err = handler.BindJSONBody(request, resp, h.Validator, "#SignupRequest", &payload)
	return
}

func (h SignupHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
		return
	}

	var result *interactionflows.AuthResult
	err = db.WithTx(h.TxContext, func() (err error) {
		result, err = h.Interactions.SignupWithLoginIDPassword(
			coreauth.GetAccessKey(req.Context()).Client.ClientID(),
			payload.LoginIDs[0].Key,
			payload.LoginIDs[0].Value,
			payload.Password,
			payload.Metadata,
			payload.OnUserDuplicate,
		)
		return
	})
	if err != nil {
		handler.WriteResponse(resp, handler.APIResponse{Error: err})
		return
	}

	result.WriteResponse(resp)
}
