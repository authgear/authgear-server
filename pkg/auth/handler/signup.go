package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
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

func AttachSignupHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/signup").
		Handler(server.FactoryToHandler(&SignupHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type SignupHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f SignupHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SignupHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
			"minItems": 1
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
	RequireAuthz         handler.RequireAuthz  `dependency:"RequireAuthz"`
	Validator            *validation.Validator `dependency:"Validator"`
	AuthnSignupProvider  authn.SignupProvider  `dependency:"AuthnSignupProvider"`
	AuthnSessionProvider authnsession.Provider `dependency:"AuthnSessionProvider"`
	TxContext            db.TxContext          `dependency:"TxContext"`
	Logger               *logrus.Entry         `dependency:"HandlerLogger"`
	TaskQueue            async.Queue           `dependency:"AsyncTaskQueue"`
	HookProvider         hook.Provider         `dependency:"HookProvider"`
}

func (h SignupHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.RequireClient)
}

func (h SignupHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload SignupRequestPayload, err error) {
	err = handler.BindJSONBody(request, resp, h.Validator, "#SignupRequest", &payload)
	return
}

func (h SignupHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
		h.AuthnSessionProvider.WriteResponse(resp, nil, err)
		return
	}

	var tasks []async.TaskSpec
	h.TxContext.UseHook(h.HookProvider)
	result, err := handler.Transactional(h.TxContext, func() (result interface{}, err error) {
		result, tasks, err = h.Handle(payload)
		return
	})
	if err == nil {
		for _, t := range tasks {
			h.TaskQueue.Enqueue(t.Name, t.Param, nil)
		}
	}
	h.AuthnSessionProvider.WriteResponse(resp, result, err)
}

func (h SignupHandler) Handle(payload SignupRequestPayload) (resp interface{}, tasks []async.TaskSpec, err error) {
	authInfo, _, firstPrincipal, tasks, err := h.AuthnSignupProvider.CreateUserWithLoginIDs(
		payload.LoginIDs,
		payload.Password,
		payload.Metadata,
		payload.OnUserDuplicate,
	)
	if err != nil {
		return
	}
	sess, err := h.AuthnSessionProvider.NewFromScratch(authInfo.ID, firstPrincipal, coreAuth.SessionCreateReasonSignup)
	if err != nil {
		return
	}
	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*sess)
	if err != nil {
		return
	}

	return
}
