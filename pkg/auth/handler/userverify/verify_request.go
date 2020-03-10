package userverify

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// AttachVerifyRequestHandler attaches VerifyRequestHandler to server
func AttachVerifyRequestHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.NewRoute().
		Path("/verify_request").
		Handler(server.FactoryToHandler(&VerifyRequestHandlerFactory{
			authDependency,
		})).
		Methods("OPTIONS", "POST")
}

// VerifyRequestHandlerFactory creates VerifyRequestHandler
type VerifyRequestHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new VerifyRequestHandler
func (f VerifyRequestHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &VerifyRequestHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type loginIDType string

const (
	loginIDTypeEmail loginIDType = "email"
	loginIDTypePhone loginIDType = "phone"
)

func (t loginIDType) MatchPrincipal(principal *password.Principal, provider password.Provider) bool {
	switch t {
	case loginIDTypeEmail:
		return provider.CheckLoginIDKeyType(principal.LoginIDKey, metadata.Email)
	case loginIDTypePhone:
		return provider.CheckLoginIDKeyType(principal.LoginIDKey, metadata.Phone)
	default:
		return false
	}
}

type VerifyRequestPayload struct {
	LoginIDType loginIDType `json:"login_id_type"`
	LoginID     string      `json:"login_id"`
}

// @JSONSchema
const VerifyRequestSchema = `
{
	"$id": "#VerifyRequest",
	"oneOf": [
		{
			"type": "object",
			"properties": {
				"login_id_type": { "enum": ["phone"] },
				"login_id": { "type": "string", "format": "phone" }
			},
			"required": ["login_id_type", "login_id"]
		},
		{
			"type": "object",
			"properties": {
				"login_id_type": { "enum": ["email"] },
				"login_id": { "type": "string", "format": "email" }
			},
			"required": ["login_id_type", "login_id"]
		}
	]
}
`

/*
	@Operation POST /verify_request - Request verification
		Request verification code to be sent to login ID.

		@Tag User Verification
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {VerifyRequest}

		@Response 200 {EmptyResponse}
*/
type VerifyRequestHandler struct {
	TxContext                db.TxContext                 `dependency:"TxContext"`
	Validator                *validation.Validator        `dependency:"Validator"`
	AuthContext              coreAuth.ContextGetter       `dependency:"AuthContextGetter"`
	RequireAuthz             handler.RequireAuthz         `dependency:"RequireAuthz"`
	CodeSenderFactory        userverify.CodeSenderFactory `dependency:"UserVerifyCodeSenderFactory"`
	URLPrefix                *url.URL                     `dependency:"URLPrefix"`
	UserVerificationProvider userverify.Provider          `dependency:"UserVerificationProvider"`
	UserProfileStore         userprofile.Store            `dependency:"UserProfileStore"`
	PasswordAuthProvider     password.Provider            `dependency:"PasswordAuthProvider"`
	IdentityProvider         principal.IdentityProvider   `dependency:"IdentityProvider"`
	Logger                   *logrus.Entry                `dependency:"HandlerLogger"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h VerifyRequestHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.RequireValidUser
}

func (h VerifyRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h VerifyRequestHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload VerifyRequestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#VerifyRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() (err error) {
		authInfo, _ := h.AuthContext.AuthInfo()

		// Get Profile
		var userProfile userprofile.UserProfile
		if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
			return
		}

		// We don't check realms. i.e. Verifying a email means every email login IDs
		// of that email is verified, regardless the realm.
		principals, err := h.PasswordAuthProvider.GetPrincipalsByLoginID("", payload.LoginID)
		if err != nil {
			return
		}

		var userPrincipal *password.Principal
		for _, principal := range principals {
			if principal.UserID == authInfo.ID && payload.LoginIDType.MatchPrincipal(principal, h.PasswordAuthProvider) {
				userPrincipal = principal
				break
			}
		}
		if userPrincipal == nil {
			err = validation.NewValidationFailed("invalid request body", []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Pointer: "/login_id",
				Message: "login ID is not owned by this user",
			}})
			return
		}

		verifyCode, err := h.UserVerificationProvider.CreateVerifyCode(userPrincipal)
		if err != nil {
			return
		}

		codeSender := h.CodeSenderFactory.NewCodeSender(h.URLPrefix, userPrincipal.LoginIDKey)
		user := model.NewUser(*authInfo, userProfile)
		if err = codeSender.Send(*verifyCode, user); err != nil {
			return
		}

		resp = struct{}{}
		return
	})
	return
}
