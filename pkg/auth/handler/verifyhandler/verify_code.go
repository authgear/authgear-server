package verifyhandler

import (
	"encoding/json"
	"fmt"
	"net/http"

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

// AttachVerifyCodeHandler attaches VerifyCodeHandler to server
func AttachVerifyCodeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/verify_code", &VerifyCodeHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
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
	TxContext        db.TxContext           `dependency:"TxContext"`
	AuthContext      coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	UserProfileStore userprofile.Store      `dependency:"UserProfileStore"`
	VerifyCodeStore  verifycode.Store       `dependency:"VerifyCodeStore"`
	Logger           *logrus.Entry          `dependency:"HandlerLogger"`
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
	code := verifycode.VerifyCode{}

	if err = h.VerifyCodeStore.GetVerifyCodeByCode(payload.Code, &code); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"code":  payload.Code,
			"error": err,
		}).Error("failed to get verify code")
		err = h.invalidCodeError(payload.Code)
		return
	}
	if code.Consumed {
		h.Logger.WithField("code", payload.Code).Error("code has been consumed")
		err = h.invalidCodeError(payload.Code)
		return
	}

	// TODO: update user verification state
	authInfo := h.AuthContext.AuthInfo()
	token := h.AuthContext.Token()
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID, token.AccessToken); err != nil {
		h.Logger.WithField("user_id", authInfo.ID).Error("unexpected user not found")
		err = skyerr.NewError(skyerr.UnexpectedUserNotFound, "user not found")
		return
	}

	if userProfile.Data[code.RecordKey] != code.RecordValue {
		err = skyerr.NewError(
			skyerr.InvalidArgument,
			"the user data has since been modified, a new verification is required",
		)
		return
	}

	resp = "OK"
	return
}

func (h VerifyCodeHandler) invalidCodeError(code string) error {
	msg := fmt.Sprintf(
		"the code `%s` is not valid for user `%s`",
		code, h.AuthContext.AuthInfo().ID,
	)
	return skyerr.NewInvalidArgument(msg, []string{"code"})
}
