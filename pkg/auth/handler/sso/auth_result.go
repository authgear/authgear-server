package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func AttachAuthResultHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/auth_result").
		Handler(server.FactoryToHandler(&AuthResultHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type AuthResultHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f *AuthResultHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &AuthResultHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type AuthResultHandler struct {
	TxContext            db.TxContext          `dependency:"TxContext"`
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	HookProvider         hook.Provider         `dependency:"HookProvider"`
	AuthnSessionProvider authnsession.Provider `dependency:"AuthnSessionProvider"`
	AuthnOAuthProvider   authn.OAuthProvider   `dependency:"AuthnOAuthProvider"`
	Validator            *validation.Validator `dependency:"Validator"`
	TaskQueue            async.Queue           `dependency:"AsyncTaskQueue"`
	SSOProvider          sso.Provider          `dependency:"SSOProvider"`
}

func (h *AuthResultHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

type AuthResultPayload struct {
	AuthorizationCode string `json:"authorization_code"`
	CodeVerifier      string `json:"code_verifier"`
}

// @JSONSchema
const AuthResultRequestSchema = `
{
	"$id": "#AuthResultRequest",
	"type": "object",
	"properties": {
		"authorization_code": { "type": "string", "minLength": 1 },
		"code_verifier": { "type": "string", "minLength": 1 }
	},
	"required": ["authorization_code", "code_verifier"]
}
`

func (h *AuthResultHandler) DecodeRequest(w http.ResponseWriter, r *http.Request) (payload *AuthResultPayload, err error) {
	err = handler.BindJSONBody(r, w, h.Validator, "#AuthResultRequest", &payload)
	return
}

func (h *AuthResultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var result interface{}
	var err error

	payload, err := h.DecodeRequest(w, r)
	if err != nil {
		h.AuthnSessionProvider.WriteResponse(w, nil, err)
		return
	}

	err = db.WithTx(h.TxContext, func() (err error) {
		result, err = h.Handle(payload)
		return
	})
	h.AuthnSessionProvider.WriteResponse(w, result, err)
}

func (h *AuthResultHandler) Handle(payload *AuthResultPayload) (result interface{}, err error) {
	code, err := h.SSOProvider.DecodeSkygearAuthorizationCode(payload.AuthorizationCode)
	if err != nil {
		return
	}

	err = h.SSOProvider.VerifyPKCE(code, payload.CodeVerifier)
	if err != nil {
		return
	}

	authInfo, userProfile, prin, err := h.AuthnOAuthProvider.ExtractAuthorizationCode(code)
	if err != nil {
		return
	}

	if code.Action == "link" {
		user := model.NewUser(*authInfo, *userProfile)
		result = model.NewAuthResponseWithUser(user)
		return
	}

	// code.Action == "login"
	sess, err := h.AuthnSessionProvider.NewFromScratch(code.UserID, prin, coreAuth.SessionCreateReason(code.SessionCreateReason))
	if err != nil {
		return
	}

	result, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*sess)
	if err != nil {
		return
	}

	return
}
