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
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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
	return h.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h)
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
		"email": { "type": "string" }
	}
}
`

func (payload ForgotPasswordPayload) Validate() error {
	// TODO(error): JSON schema
	if payload.Email == "" {
		return skyerr.NewInvalid("empty email")
	}

	return nil
}

/*
	@Operation POST /forgot_password - Request password recovery
		Request password recovery message to be sent to email.

		@Tag Forgot Password

		@RequestBody
			@JSONSchema {ForgotPasswordRequest}

		@Response 200 {EmptyResponse}
*/
type ForgotPasswordHandler struct {
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

func (h ForgotPasswordHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h ForgotPasswordHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := ForgotPasswordPayload{}
	if err := handler.DecodeJSONBody(request, resp, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h ForgotPasswordHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ForgotPasswordPayload)

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

	resp = map[string]string{}
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
	return h.RequireAuthz(handler.APIHandlerToHandler(h, nil), h)
}

type ForgotPasswordTestPayload struct {
	Email        string `json:"email"`
	TextTemplate string `json:"text_template"`
	HTMLTemplate string `json:"html_template"`
	Subject      string `json:"subject"`
	Sender       string `json:"sender"`
	ReplyTo      string `json:"reply_to"`
}

func (payload ForgotPasswordTestPayload) Validate() error {
	// TODO(error): JSON schema
	if payload.Email == "" {
		return skyerr.NewInvalid("empty email")
	}

	return nil
}

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
	RequireAuthz              handler.RequireAuthz `dependency:"RequireAuthz"`
	ForgotPasswordEmailSender welcemail.TestSender `dependency:"TestForgotPasswordEmailSender"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h ForgotPasswordTestHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h ForgotPasswordTestHandler) WithTx() bool {
	return false
}

// DecodeRequest decode request payload
func (h ForgotPasswordTestHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := ForgotPasswordTestPayload{}
	if err := handler.DecodeJSONBody(request, resp, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h ForgotPasswordTestHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ForgotPasswordTestPayload)
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
