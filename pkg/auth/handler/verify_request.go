package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify/verifycode"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AttachVerifyRequestHandler attaches VerifyRequestHandler to server
func AttachVerifyRequestHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/verify_request", &VerifyRequestHandlerFactory{
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
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f VerifyRequestHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	// FIXME: Admin only after adding admin role
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type VerifyRequestPayload struct {
	RecordKey string `json:"record_key"`
}

func (payload VerifyRequestPayload) Validate() error {
	if payload.RecordKey == "" {
		return skyerr.NewInvalidArgument("empty record_key", []string{"record_key"})
	}

	return nil
}

// VerifyRequestHandler allows client to request verification (i.e. send email or send SMS).
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/verify_request <<EOF
//  {
//    "record_key": "email"
//  }
//  EOF
//
type VerifyRequestHandler struct {
	TxContext         db.TxContext                     `dependency:"TxContext"`
	AuthContext       coreAuth.ContextGetter           `dependency:"AuthContextGetter"`
	CodeSenderFactory auth.UserVerifyCodeSenderFactory `dependency:"UserVerifyCodeSenderFactory"`
	UserProfileStore  userprofile.Store                `dependency:"UserProfileStore"`
	VerifyCodeStore   verifycode.Store                 `dependency:"VerifyCodeStore"`
	Logger            *logrus.Entry                    `dependency:"HandlerLogger"`
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
	accessToken := h.AuthContext.Token()
	codeSender := h.CodeSenderFactory.NewCodeSender(payload.RecordKey)
	if codeSender == nil {
		err = skyerr.NewInvalidArgument("invalid record_key", []string{payload.RecordKey})
	}

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID, accessToken.AccessToken); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to fetch user profile")
		return
	}

	var value string
	var ok bool
	if value, ok = userProfile.Data[payload.RecordKey].(string); !ok {
		err = skyerr.NewError(skyerr.UnexpectedError, "Value of "+payload.RecordKey+" is not string")
		return
	}

	code := codeSender.Generate()
	if err = codeSender.Send(code, payload.RecordKey, value, userProfile); err != nil {
		h.Logger.WithFields(logrus.Fields{
			"error":        err,
			"record_key":   payload.RecordKey,
			"record_value": value,
		}).Error("fail to send verify request")
		return
	}

	verifyCode := verifycode.NewVerifyCode()
	verifyCode.UserID = authInfo.ID
	verifyCode.RecordKey = payload.RecordKey
	verifyCode.RecordValue = value
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = time.Now()

	if err = h.VerifyCodeStore.CreateVerifyCode(&verifyCode); err != nil {
		return
	}

	resp = "OK"
	return
}
