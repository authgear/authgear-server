package forgotpwd

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// AttachForgotPasswordHandler attaches ForgotPasswordHandler to server
func AttachForgotPasswordHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/forgot_password", &ForgotPasswordHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ForgotPasswordHandlerFactory creates ForgotPasswordHandler
type ForgotPasswordHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordHandler
func (f ForgotPasswordHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
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
	Validator                 *validation.Validator      `dependency:"Validator"`
	TxContext                 db.TxContext               `dependency:"TxContext"`
	RequireAuthz              handler.RequireAuthz       `dependency:"RequireAuthz"`
	ForgotPasswordEmailSender forgotpwdemail.Sender      `dependency:"ForgotPasswordEmailSender"`
	PasswordAuthProvider      password.Provider          `dependency:"PasswordAuthProvider"`
	IdentityProvider          principal.IdentityProvider `dependency:"IdentityProvider"`
	AuthInfoStore             authinfo.Store             `dependency:"AuthInfoStore"`
	UserProfileStore          userprofile.Store          `dependency:"UserProfileStore"`
	SecureMatch               bool                       `dependency:"ForgotPasswordSecureMatch"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h ForgotPasswordHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

func (h ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h ForgotPasswordHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ForgotPasswordPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#ForgotPasswordRequest", &payload); err != nil {
		return nil, err
	}

	err = db.WithTx(h.TxContext, func() (err error) {
		principals, err := h.PasswordAuthProvider.GetPrincipalsByLoginID("", payload.Email)
		if err != nil {
			return
		}

		principalMap := map[string]*password.Principal{}
		for _, principal := range principals {
			if h.PasswordAuthProvider.CheckLoginIDKeyType(principal.LoginIDKey, metadata.Email) {
				principalMap[principal.UserID] = principal
			}
		}

		if len(principalMap) == 0 {
			if h.SecureMatch {
				resp = map[string]string{}
			} else {
				err = authinfo.ErrNotFound
			}
			return
		}

		for userID, principal := range principalMap {
			hashedPassword := principal.HashedPassword

			fetchedAuthInfo := authinfo.AuthInfo{}
			if err = h.AuthInfoStore.GetAuth(userID, &fetchedAuthInfo); err != nil {
				return
			}

			// Get Profile
			var userProfile userprofile.UserProfile
			if userProfile, err = h.UserProfileStore.GetUserProfile(fetchedAuthInfo.ID); err != nil {
				return
			}

			user := model.NewUser(fetchedAuthInfo, userProfile)

			if err = h.ForgotPasswordEmailSender.Send(
				payload.Email,
				fetchedAuthInfo,
				user,
				hashedPassword,
			); err != nil {
				return
			}
		}

		return
	})

	resp = struct{}{}
	return
}
