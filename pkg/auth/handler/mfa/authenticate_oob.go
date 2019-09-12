package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachAuthenticateOOBHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/oob/authenticate", &AuthenticateOOBHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

type AuthenticateOOBHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f AuthenticateOOBHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthenticateOOBHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.RequireAuthz(h, h.AuthContext, h)
}

type AuthenticateOOBRequest struct {
	AuthnSessionToken  string `json:"authn_session_token"`
	Code               string `json:"code"`
	RequestBearerToken bool   `json:"request_bearer_token"`
}

func (r AuthenticateOOBRequest) Validate() error {
	if r.AuthnSessionToken == "" {
		return skyerr.NewInvalidArgument("missing authentication session token", []string{"authn_session_token"})
	}
	if r.Code == "" {
		return skyerr.NewInvalidArgument("missing code", []string{"code"})
	}
	return nil
}

// @JSONSchema
const AuthenticateOOBRequestSchema = `
{
	"$id": "#AuthenticateOOBRequest",
	"type": "object",
	"properties": {
		"authn_session_token": { "type": "string" },
		"code": { "type": "string" },
		"request_bearer_token": { "type": "boolean" }
	}
	"required": ["authn_session_token", "code"]
}
`

/*
	@Operation POST /mfa/oob/authenticate - Authenticate with OOB.
		Authenticate with OOB.

		@Tag User
		@SecurityRequirement access_key

		@RequestBody {AuthenticateOOBRequest}
		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type AuthenticateOOBHandler struct {
	TxContext            db.TxContext            `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter  `dependency:"AuthContextGetter"`
	MFAProvider          mfa.Provider            `dependency:"MFAProvider"`
	MFAConfiguration     config.MFAConfiguration `dependency:"MFAConfiguration"`
	HookProvider         hook.Provider           `dependency:"HookProvider"`
	AuthnSessionProvider authnsession.Provider   `dependency:"AuthnSessionProvider"`
}

func (h *AuthenticateOOBHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(authz.PolicyFunc(policy.DenyNoAccessKey))
}

func (h *AuthenticateOOBHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := AuthenticateOOBRequest{}
	err := handler.DecodeJSONBody(request, &payload)
	return payload, err
}

func (h *AuthenticateOOBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	defer func() {
		if err == nil {
			h.HookProvider.DidCommitTx()
		}
		h.AuthnSessionProvider.WriteResponse(w, result, err)
	}()

	payload, err := h.DecodeRequest(r)
	if err != nil {
		return
	}

	err = payload.Validate()
	if err != nil {
		return
	}

	result, err = handler.Transactional(h.TxContext, func() (result interface{}, err error) {
		result, err = h.Handle(payload)
		if err == nil {
			err = h.HookProvider.WillCommitTx()
		}
		return
	})
}

func (h *AuthenticateOOBHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(AuthenticateOOBRequest)

	authnSess, err := h.AuthnSessionProvider.NewWithToken(payload.AuthnSessionToken)
	if err != nil {
		return
	}

	a, bearerToken, err := h.MFAProvider.AuthenticateOOB(authnSess.UserID, payload.Code, payload.RequestBearerToken)
	if err != nil {
		return
	}

	err = authnSess.StepMFA(coreAuth.AuthnSessionStepMFAOptions{
		AuthenticatorID:          a.ID,
		AuthenticatorType:        a.Type,
		AuthenticatorOOBChannel:  a.Channel,
		AuthenticatorBearerToken: bearerToken,
	})
	if err != nil {
		return
	}

	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*authnSess)
	if err != nil {
		return
	}

	return
}
