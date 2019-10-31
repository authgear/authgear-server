package forgotpwd

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"
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
	server.Handle("/forgot_password/test", &ForgotPasswordTestHandlerFactory{
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
	}
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

// ForgotPasswordTestHandlerFactory creates ForgotPasswordTestHandler
type ForgotPasswordTestHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordTestHandler
func (f ForgotPasswordTestHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordTestHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type ForgotPasswordTestPayload struct {
	Email        string `json:"email"`
	TextTemplate string `json:"text_template"`
	HTMLTemplate string `json:"html_template"`
	Subject      string `json:"subject"`
	Sender       string `json:"sender"`
	ReplyTo      string `json:"reply_to"`
}

// nolint: gosec
const ForgotPasswordTestRequestSchema = `
{
	"$id": "#ForgotPasswordTestRequest",
	"type": "object",
	"properties": {
		"email": { "type": "string", "format": "email" },
		"text_template": { "type": "string", "minLength": 1 },
		"html_template": { "type": "string", "minLength": 1 },
		"subject": { "type": "string", "minLength": 1 },
		"sender": { "type": "string", "minLength": 1 },
		"reply_to": { "type": "string", "minLength": 1 }
	}
}
`

// ForgotPasswordTestHandler send a dummy reset password email to given email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/forgot_password/test <<EOF
//  {
//     "email": "xxx@oursky.com",
//     "text_template": "xxx",
//     "html_template": "xxx",
//     "subject": "xxx",
//     "sender": "xxx",
//     "reply_to": "xxx"
//  }
//  EOF
type ForgotPasswordTestHandler struct {
	RequireAuthz              handler.RequireAuthz  `dependency:"RequireAuthz"`
	Validator                 *validation.Validator `dependency:"Validator"`
	ForgotPasswordEmailSender welcemail.TestSender  `dependency:"TestForgotPasswordEmailSender"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h ForgotPasswordTestHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h ForgotPasswordTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h ForgotPasswordTestHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload ForgotPasswordTestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#ForgotPasswordTestRequest", &payload); err != nil {
		return nil, err
	}

	if err = h.ForgotPasswordEmailSender.Send(
		payload.Email,
		payload.TextTemplate,
		payload.HTMLTemplate,
		payload.Subject,
		payload.Sender,
		payload.ReplyTo,
	); err == nil {
		resp = map[string]string{}
	}

	return
}
