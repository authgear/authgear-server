package userverify

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/utils"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
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
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

// AttachVerifyRequestHandler attaches VerifyRequestHandler to server
func AttachVerifyRequestHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/verify_request", &VerifyRequestHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	server.Handle("/verify_request/test", &VerifyRequestTestHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// VerifyRequestHandlerFactory creates VerifyRequestHandler
type VerifyRequestHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new VerifyRequestHandler
func (f VerifyRequestHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &VerifyRequestHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(handler.APIHandlerToHandler(h, h.TxContext), h)
}

type loginIDType string

const (
	loginIDTypeEmail loginIDType = "email"
	loginIDTypePhone loginIDType = "phone"
)

var allLoginIDTypes = []string{
	string(loginIDTypeEmail),
	string(loginIDTypePhone),
}

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
	"type": "object",
	"properties": {
		"login_id_type": { "type": "string" },
		"login_id": { "type": "string" }
	}
}
`

func (payload VerifyRequestPayload) Validate() error {
	// TODO(error): JSON schema
	if !utils.StringSliceContains(allLoginIDTypes, string(payload.LoginIDType)) {
		return skyerr.NewInvalid("invalid login ID type")
	}

	if payload.LoginID == "" {
		return skyerr.NewInvalid("empty login ID")
	}

	return nil
}

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
	AuthContext              coreAuth.ContextGetter       `dependency:"AuthContextGetter"`
	RequireAuthz             handler.RequireAuthz         `dependency:"RequireAuthz"`
	CodeSenderFactory        userverify.CodeSenderFactory `dependency:"UserVerifyCodeSenderFactory"`
	UserVerificationProvider userverify.Provider          `dependency:"UserVerificationProvider"`
	UserProfileStore         userprofile.Store            `dependency:"UserProfileStore"`
	PasswordAuthProvider     password.Provider            `dependency:"PasswordAuthProvider"`
	IdentityProvider         principal.IdentityProvider   `dependency:"IdentityProvider"`
	Logger                   *logrus.Entry                `dependency:"HandlerLogger"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h VerifyRequestHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

func (h VerifyRequestHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h VerifyRequestHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := VerifyRequestPayload{}
	if err := handler.DecodeJSONBody(request, resp, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h VerifyRequestHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(VerifyRequestPayload)
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
		// TODO(error): JSON schema
		err = skyerr.NewInvalid("login ID is not owned by this user")
		return
	}

	verifyCode, err := h.UserVerificationProvider.CreateVerifyCode(userPrincipal)
	if err != nil {
		return
	}

	codeSender := h.CodeSenderFactory.NewCodeSender(userPrincipal.LoginIDKey)
	user := model.NewUser(*authInfo, userProfile)
	if err = codeSender.Send(*verifyCode, user); err != nil {
		return
	}

	resp = map[string]string{}
	return
}

// VerifyRequestTestHandlerFactory creates VerifyRequestTestHandler
type VerifyRequestTestHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new VerifyRequestTestHandler
func (f VerifyRequestTestHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &VerifyRequestTestHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(handler.APIHandlerToHandler(h, nil), h)
}

type VerifyRequestTestPayload struct {
	LoginIDKey    string                                       `json:"login_id_key"`
	LoginID       string                                       `json:"login_id"`
	User          model.User                                   `json:"user"`
	MessageConfig config.UserVerificationProviderConfiguration `json:"message_config"`
	Templates     map[string]string                            `json:"templates"`
}

func (payload VerifyRequestTestPayload) Validate() error {
	// TODO(error): JSON schema
	if payload.LoginIDKey == "" {
		return skyerr.NewInvalid("empty login_id_key")
	}

	if payload.LoginID == "" {
		return skyerr.NewInvalid("empty login_id")
	}

	return nil
}

// VerifyRequestTestHandler sends a dummy verification request (i.e. email or SMS).
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/verify_request/test <<EOF
//  {
//    "login_id_key": "email",
//    "login_id": "user@example.com",
//    "user":  {
//      "user_id": "5bf1e4d2-e1c4-4517-93c7-7dae89261da6",
//      "metadata": {
//        "email": "user@example.com"
//      }
//    },
//    "templates": {
//      "text": "testing",
//      "html": "testing html"
//    },
//    "message_config": {
//      "subject": "Test"
//    },
//  }
//  EOF
//
type VerifyRequestTestHandler struct {
	RequireAuthz          handler.RequireAuthz             `dependency:"RequireAuthz"`
	TestCodeSenderFactory userverify.TestCodeSenderFactory `dependency:"UserVerifyTestCodeSenderFactory"`
	Logger                *logrus.Entry                    `dependency:"HandlerLogger"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h VerifyRequestTestHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h VerifyRequestTestHandler) WithTx() bool {
	return false
}

// DecodeRequest decode request payload
func (h VerifyRequestTestHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (handler.RequestPayload, error) {
	payload := VerifyRequestTestPayload{}
	if err := handler.DecodeJSONBody(request, resp, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (h VerifyRequestTestHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(VerifyRequestTestPayload)
	sender := h.TestCodeSenderFactory.NewTestCodeSender(
		payload.MessageConfig,
		payload.LoginIDKey,
		payload.Templates,
	)

	user := payload.User
	if user.ID == "" {
		user.ID = "test-user-id"
	}

	verifyCode := userverify.NewVerifyCode()
	verifyCode.UserID = user.ID
	verifyCode.LoginIDKey = payload.LoginIDKey
	verifyCode.LoginID = payload.LoginID
	verifyCode.Code = "TEST1234"
	verifyCode.Consumed = false
	verifyCode.CreatedAt = time.Now().UTC()

	err = sender.Send(verifyCode, user)
	if err != nil {
		return
	}

	resp = map[string]string{}

	return
}
