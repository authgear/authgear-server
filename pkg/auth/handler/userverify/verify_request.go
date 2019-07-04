package userverify

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/utils"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f VerifyRequestHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type loginIDType string

const (
	loginIDTypeEmail loginIDType = "email"
)

var allLoginIDTypes = []string{
	string(loginIDTypeEmail),
}

func (t loginIDType) MatchPrincipal(principal *password.Principal, provider password.Provider) bool {
	switch t {
	case loginIDTypeEmail:
		return provider.CheckLoginIDKeyType(principal.LoginIDKey, metadata.Email)
	default:
		return false
	}
}

type VerifyRequestPayload struct {
	LoginIDType loginIDType `json:"login_id_type"`
	LoginID     string      `json:"login_id"`
}

func (payload VerifyRequestPayload) Validate() error {
	if !utils.StringSliceContains(allLoginIDTypes, string(payload.LoginIDType)) {
		return skyerr.NewInvalidArgument("invalid login ID type", []string{"login_id_type"})
	}

	if payload.LoginID == "" {
		return skyerr.NewInvalidArgument("empty login ID", []string{"login_id"})
	}

	return nil
}

// VerifyRequestHandler allows client to request verification (i.e. send email or send SMS).
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/verify_request <<EOF
//  {
//    "login_id_type": "email",
//    "login_id": "user@example.com"
//  }
//  EOF
//
type VerifyRequestHandler struct {
	TxContext            db.TxContext                    `dependency:"TxContext"`
	AuthContext          coreAuth.ContextGetter          `dependency:"AuthContextGetter"`
	CodeSenderFactory    userverify.CodeSenderFactory    `dependency:"UserVerifyCodeSenderFactory"`
	CodeGeneratorFactory userverify.CodeGeneratorFactory `dependency:"VerifyCodeCodeGeneratorFactory"`
	UserProfileStore     userprofile.Store               `dependency:"UserProfileStore"`
	VerifyCodeStore      userverify.Store                `dependency:"VerifyCodeStore"`
	Logger               *logrus.Entry                   `dependency:"HandlerLogger"`
	PasswordAuthProvider password.Provider               `dependency:"PasswordAuthProvider"`
}

func (h VerifyRequestHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h VerifyRequestHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := VerifyRequestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

func (h VerifyRequestHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(VerifyRequestPayload)
	authInfo := h.AuthContext.AuthInfo()

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	userFactory := response.UserFactory{
		PasswordAuthProvider: h.PasswordAuthProvider,
	}
	user := userFactory.NewUser(*authInfo, userProfile)

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
		err = skyerr.NewError(skyerr.UnexpectedError, "Value of "+payload.LoginID+" doesn't exist.")
		return
	}

	codeSender := h.CodeSenderFactory.NewCodeSender(userPrincipal.LoginIDKey)
	if codeSender == nil {
		err = skyerr.NewError(skyerr.UnexpectedError, "User verification not configured for login ID key")
		return
	}

	codeGenerator := h.CodeGeneratorFactory.NewCodeGenerator(userPrincipal.LoginIDKey)
	code := codeGenerator.Generate()

	verifyCode := userverify.NewVerifyCode()
	verifyCode.UserID = authInfo.ID
	verifyCode.RecordKey = userPrincipal.LoginIDKey
	verifyCode.RecordValue = userPrincipal.LoginID
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = time.Now()

	if err = h.VerifyCodeStore.CreateVerifyCode(&verifyCode); err != nil {
		return
	}

	if err = codeSender.Send(verifyCode, user); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":        err,
			"login_id_key": userPrincipal.LoginIDKey,
			"login_id":     userPrincipal.LoginID,
		}).Error("fail to send verify request")
		return
	}

	resp = "OK"
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
	return handler.APIHandlerToHandler(h, nil)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f VerifyRequestTestHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.RequireMasterKey),
	)
}

type VerifyRequestTestPayload struct {
	RecordKey        string            `json:"record_key"`
	RecordValue      string            `json:"record_value"`
	ProviderSettings map[string]string `json:"provider_settings"`
	Templates        map[string]string `json:"templates"`
}

func (payload VerifyRequestTestPayload) Validate() error {
	if payload.RecordKey == "" {
		return skyerr.NewInvalidArgument("empty record_key", []string{"record_key"})
	}

	if payload.RecordValue == "" {
		return skyerr.NewInvalidArgument("empty record_value", []string{"record_value"})
	}

	if payload.ProviderSettings == nil || payload.ProviderSettings["name"] == "" {
		return skyerr.NewInvalidArgument("missing provider name", []string{"provider_settings.name"})
	}

	return nil
}

// VerifyRequestTestHandler sends a dummy verification request (i.e. email or SMS).
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/verify_request/test <<EOF
//  {
//    "record_key": "email",
//    "record_value": "test@example.com",
//    "provider_settings": {
//      "name": "smtp"
//    },
//    "templates": {
//      "text": "testing",
//      "html": "testing html"
//    }
//  }
//  EOF
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/verify_request/test <<EOF
//  {
//    "record_key": "phone",
//    "record_value": "+15005550009",
//    "provider_settings": {
//      "name": "twilio",
//      "twilio_from": "+15005550009",
//      "twilio_account_sid": "",
//      "twilio_auth_token": ""
//    },
//    "templates": {
//      "text": "testing sms"
//    }
//  }
//  EOF
//
type VerifyRequestTestHandler struct {
	TestCodeSenderFactory userverify.TestCodeSenderFactory `dependency:"UserVerifyTestCodeSenderFactory"`
	Logger                *logrus.Entry                    `dependency:"HandlerLogger"`
}

func (h VerifyRequestTestHandler) WithTx() bool {
	return false
}

// DecodeRequest decode request payload
func (h VerifyRequestTestHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := VerifyRequestTestPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

func (h VerifyRequestTestHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(VerifyRequestTestPayload)
	codeSender := h.TestCodeSenderFactory.NewTestCodeSender(payload.RecordKey, payload.ProviderSettings, payload.Templates)
	if codeSender == nil {
		err = skyerr.NewInvalidArgument("invalid provider name", []string{"provider_settings.name"})
		return
	}

	if err = codeSender.Send(payload.RecordKey, payload.RecordValue); err == nil {
		resp = "OK"
	}

	return
}
