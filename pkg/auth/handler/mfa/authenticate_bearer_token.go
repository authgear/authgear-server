package mfa

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthenticateBearerTokenHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/mfa/bearer_token/authenticate").
		Handler(auth.MakeHandler(authDependency, newAuthenticateBearerTokenHandler)).
		Methods("OPTIONS", "POST")
}

type AuthenticateBearerTokenRequest struct {
	AuthnSessionToken string `json:"authn_session_token"`
	BearerToken       string `json:"bearer_token"`
}

func (r *AuthenticateBearerTokenRequest) Validate() []validation.ErrorCause {
	if len(r.BearerToken) > 0 {
		return nil
	}
	return []validation.ErrorCause{{
		Kind:    validation.ErrorRequired,
		Pointer: "/bearer_token",
		Message: "bearer_token is required",
	}}
}

// nolint: gosec
// @JSONSchema
const AuthenticateBearerTokenRequestSchema = `
{
	"$id": "#AuthenticateBearerTokenRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string", "minLength": 1 },
		"bearer_token": { "type": "string", "minLength": 1 }
	}
}
`

/*
	@Operation POST /mfa/bearer_token/authenticate - Authenticate with bearer token.
		Authenticate with bearer token.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody
			@JSONSchema {AuthenticateBearerTokenRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateBearerTokenHandler struct {
	TxContext         db.TxContext
	Validator         *validation.Validator
	TimeProvider      time.Provider
	MFAProvider       mfa.Provider
	BearerTokenCookie mfa.BearerTokenCookieConfiguration
	authnResolver     authnResolver
	authnStepper      authnStepper
}

func (h *AuthenticateBearerTokenHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

func (h *AuthenticateBearerTokenHandler) useCookie(r *http.Request) bool {
	accessKey := coreAuth.GetAccessKey(r.Context())
	return accessKey.Client == nil || accessKey.Client.AuthAPIUseCookie()
}

func (h *AuthenticateBearerTokenHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload AuthenticateBearerTokenRequest, err error) {
	if h.useCookie(request) {
		cookie, err := request.Cookie(h.BearerTokenCookie.Name)
		if err == nil {
			payload.BearerToken = cookie.Value
		}
	}

	err = handler.BindJSONBody(request, resp, h.Validator, "#AuthenticateBearerTokenRequest", &payload)
	return
}

func (h *AuthenticateBearerTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	payload, err := h.DecodeRequest(r, w)
	if err != nil {
		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	var result authn.Result
	err = db.WithTx(h.TxContext, func() error {
		session := authn.GetSession(r.Context())
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
		attrs := session.SessionAttrs()

		err = h.MFAProvider.DeleteExpiredBearerToken(attrs.UserID)
		if err != nil {
			return err
		}

		a, err := h.MFAProvider.AuthenticateBearerToken(
			attrs.UserID,
			payload.BearerToken,
		)
		if err != nil {
			return err
		}

		now := h.TimeProvider.NowUTC()
		attrs.AuthenticatorID = a.ID
		attrs.AuthenticatorType = a.Type
		attrs.AuthenticatorUpdatedAt = &now

		result, err = h.authnStepper.StepSession(
			coreAuth.GetAccessKey(r.Context()).Client,
			session,
			"",
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if skyerr.IsKind(err, mfa.InvalidBearerToken) && h.useCookie(r) {
			h.BearerTokenCookie.Clear(w)
		}

		handler.WriteResponse(w, handler.APIResponse{Error: err})
		return
	}

	h.authnStepper.WriteResult(w, result)
}
