package userverify

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/config"
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
	TxContext                db.TxContext                 `dependency:"TxContext"`
	AuthContext              coreAuth.ContextGetter       `dependency:"AuthContextGetter"`
	CodeSenderFactory        userverify.CodeSenderFactory `dependency:"UserVerifyCodeSenderFactory"`
	UserVerificationProvider userverify.Provider          `dependency:"UserVerificationProvider"`
	UserProfileStore         userprofile.Store            `dependency:"UserProfileStore"`
	PasswordAuthProvider     password.Provider            `dependency:"PasswordAuthProvider"`
	Logger                   *logrus.Entry                `dependency:"HandlerLogger"`
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

	verifyCode, err := h.UserVerificationProvider.CreateVerifyCode(userPrincipal)
	if err != nil {
		return
	}

	codeSender := h.CodeSenderFactory.NewCodeSender(userPrincipal.LoginIDKey)
	if err = codeSender.Send(*verifyCode, user); err != nil {
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
	LoginIDKey    string                                       `json:"login_id_key"`
	LoginID       string                                       `json:"login_id"`
	User          response.User                                `json:"user"`
	Provider      string                                       `json:"provider"`
	MessageConfig config.UserVerificationProviderConfiguration `json:"message_config"`
	Templates     map[string]string                            `json:"templates"`
}

func (payload VerifyRequestTestPayload) Validate() error {
	if payload.LoginIDKey == "" {
		return skyerr.NewInvalidArgument("empty login_id_key", []string{"login_id_key"})
	}

	if payload.LoginID == "" {
		return skyerr.NewInvalidArgument("empty login_id", []string{"login_id"})
	}

	if payload.Provider == "" {
		return skyerr.NewInvalidArgument("missing provider name", []string{"provider"})
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
//    "provider": "smtp",
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
	sender := h.TestCodeSenderFactory.NewTestCodeSender(
		payload.Provider,
		payload.MessageConfig,
		payload.LoginIDKey,
		payload.Templates,
	)

	user := payload.User
	if user.UserID == "" {
		user.UserID = "test-user-id"
	}

	verifyCode := userverify.NewVerifyCode()
	verifyCode.UserID = user.UserID
	verifyCode.LoginIDKey = payload.LoginIDKey
	verifyCode.LoginID = payload.LoginID
	verifyCode.Code = "TEST1234"
	verifyCode.Consumed = false
	verifyCode.CreatedAt = time.Now().UTC()

	err = sender.Send(verifyCode, user)
	if err != nil {
		return
	}

	resp = "OK"

	return
}
