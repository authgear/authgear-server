package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	coreauthn "github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthenticateOOBHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/oob/authenticate").
		Handler(pkg.MakeHandler(authDependency, newAuthenticateOOBHandler)).
		Methods("OPTIONS", "POST")
}

type AuthenticateOOBRequest struct {
	AuthnSessionToken  string `json:"authn_session_token"`
	Code               string `json:"code"`
	RequestBearerToken bool   `json:"request_bearer_token"`
}

// @JSONSchema
const AuthenticateOOBRequestSchema = `
{
	"$id": "#AuthenticateOOBRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string", "minLength": 1 },
		"code": { "type": "string", "minLength": 1 },
		"request_bearer_token": { "type": "boolean" }
	},
	"required": ["code"]
}
`

/*
	@Operation POST /mfa/oob/authenticate - Authenticate with OOB authenticator.
		Authenticate with OOB authenticator.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody
			@JSONSchema {AuthenticateOOBRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateOOBHandler struct {
	TxContext     db.TxContext
	Validator     *validation.Validator
	TimeProvider  time.Provider
	MFAProvider   mfa.Provider
	authnResolver authnResolver
	authnStepper  authnStepper
}

func (h *AuthenticateOOBHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireClient
}

func (h *AuthenticateOOBHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (AuthenticateOOBRequest, error) {
	payload := AuthenticateOOBRequest{}
	err := handler.BindJSONBody(request, resp, h.Validator, "#AuthenticateOOBRequest", &payload)
	return payload, err
}

func (h *AuthenticateOOBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	payload, err := h.DecodeRequest(r, w)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	var result authn.Result
	err = db.WithTx(h.TxContext, func() error {
		var session coreauthn.Attributer = auth.GetSession(r.Context())
		if session == nil {
			session, err = h.authnResolver.Resolve(
				coreAuth.GetAccessKey(r.Context()).Client,
				payload.AuthnSessionToken,
				authn.SessionStep.IsMFA,
			)
			if err != nil {
				return err
			}
		}

		attrs := session.AuthnAttrs()
		a, bearerToken, err := h.MFAProvider.AuthenticateOOB(
			attrs.UserID,
			payload.Code,
			payload.RequestBearerToken,
		)
		if err != nil {
			return err
		}

		now := h.TimeProvider.NowUTC()
		attrs.AuthenticatorID = a.ID
		attrs.AuthenticatorType = a.Type
		attrs.AuthenticatorOOBChannel = a.Channel
		attrs.AuthenticatorUpdatedAt = &now

		result, err = h.authnStepper.StepSession(
			coreAuth.GetAccessKey(r.Context()).Client,
			session,
			bearerToken,
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	h.authnStepper.WriteAPIResult(w, result)
}
