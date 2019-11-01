package userverify

import (
	"net/http"
	"net/url"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/validation"

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
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		policy.RequireValidUser,
	)
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

// VerifyRequestTestHandlerFactory creates VerifyRequestTestHandler
type VerifyRequestTestHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new VerifyRequestTestHandler
func (f VerifyRequestTestHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &VerifyRequestTestHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h.RequireAuthz(h, h)
}

type VerifyRequestTestPayload struct {
	LoginIDKey    string                                       `json:"login_id_key"`
	LoginID       string                                       `json:"login_id"`
	User          model.User                                   `json:"user"`
	MessageConfig config.UserVerificationProviderConfiguration `json:"message_config"`
	Templates     map[string]string                            `json:"templates"`
}

const VerifyTestRequestSchema = `
{
	"$id": "#VerifyTestRequest",
	"type": "object",
	"properties": {
		"login_id_key": { "type": "string", "minLength": 1 },
		"login_id": { "type": "string", "minLength": 1 },
		"user": { "type": "object" },
		"html_template": { "type": "string", "minLength": 1 },
		"message_config": { "type": "object" },
		"templates": { "type": "object" }
	},
	"required": ["login_id_key", "login_id", "user", "html_template", "message_config", "templates"]
}
`

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
	Validator             *validation.Validator            `dependency:"Validator"`
	TestCodeSenderFactory userverify.TestCodeSenderFactory `dependency:"UserVerifyTestCodeSenderFactory"`
	Logger                *logrus.Entry                    `dependency:"HandlerLogger"`
}

// ProvideAuthzPolicy provides authorization policy of handler
func (h VerifyRequestTestHandler) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

func (h VerifyRequestTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	result, err := h.Handle(w, r)
	if err != nil {
		response.Error = err
	} else {
		response.Result = result
	}
	handler.WriteResponse(w, response)
}

func (h VerifyRequestTestHandler) Handle(w http.ResponseWriter, r *http.Request) (resp interface{}, err error) {
	var payload VerifyRequestTestPayload
	if err := handler.BindJSONBody(r, w, h.Validator, "#VerifyTestRequest", &payload); err != nil {
		return nil, err
	}

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

	resp = struct{}{}

	return
}
