package userverify

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AttachVerifyCodeHandler attaches VerifyCodeHandler to server
func AttachVerifyCodeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/verify_code", &VerifyCodeHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	server.Handle("/verify_code_form", &VerifyCodeFormHandlerFactory{
		authDependency,
	}).Methods("POST", "GET")
	return server
}

// VerifyCodeHandlerFactory creates VerifyCodeHandler
type VerifyCodeHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new VerifyCodeHandler
func (f VerifyCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &VerifyCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f VerifyCodeHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type VerifyCodePayload struct {
	Code string `json:"code"`
}

func (payload VerifyCodePayload) Validate() error {
	if payload.Code == "" {
		return skyerr.NewInvalidArgument("empty code", []string{"code"})
	}

	return nil
}

// VerifyCodeHandler accepts user to submit code for user verification.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/verify_code <<EOF
//  {
//    "code": "xxx"
//  }
//  EOF
//
type VerifyCodeHandler struct {
	TxContext                db.TxContext                        `dependency:"TxContext"`
	AuthContext              coreAuth.ContextGetter              `dependency:"AuthContextGetter"`
	UserProfileStore         userprofile.Store                   `dependency:"UserProfileStore"`
	VerifyCodeStore          userverify.Store                    `dependency:"VerifyCodeStore"`
	AuthInfoStore            authinfo.Store                      `dependency:"AuthInfoStore"`
	AutoUpdateUserVerifyFunc userverify.AutoUpdateUserVerifyFunc `dependency:"AutoUpdateUserVerifyFunc,optional"`
	Logger                   *logrus.Entry                       `dependency:"HandlerLogger"`
}

func (h VerifyCodeHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h VerifyCodeHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := VerifyCodePayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

func (h VerifyCodeHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(VerifyCodePayload)

	authInfo := h.AuthContext.AuthInfo()
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
		h.Logger.WithField("user_id", authInfo.ID).Error("unexpected user not found")
		err = skyerr.NewError(skyerr.UnexpectedUserNotFound, "user not found")
		return
	}

	verifyCodeReq := getAndValidateCodeRequest{
		VerifyCodeStore: h.VerifyCodeStore,
		Logger:          h.Logger,
	}

	var code userverify.VerifyCode
	if code, err = verifyCodeReq.execute(payload.Code, userProfile); err != nil {
		return
	}

	// Update code
	code.Consumed = true
	if err = h.VerifyCodeStore.UpdateVerifyCode(&code); err != nil {
		return
	}

	// Update user
	authInfo.VerifyInfo[code.RecordKey] = true
	if h.AutoUpdateUserVerifyFunc != nil {
		h.AutoUpdateUserVerifyFunc(authInfo)
	}

	if err = h.AuthInfoStore.UpdateAuth(authInfo); err != nil {
		return
	}

	resp = "OK"
	return
}
