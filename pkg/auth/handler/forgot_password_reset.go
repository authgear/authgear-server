package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// AttachForgotPasswordResetHandler attaches ForgotPasswordResetHandler to server
func AttachForgotPasswordResetHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/forgot_password/reset_password", &ForgotPasswordResetHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// ForgotPasswordResetHandlerFactory creates ForgotPasswordResetHandler
type ForgotPasswordResetHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordResetHandler
func (f ForgotPasswordResetHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordResetHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ForgotPasswordResetHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type ForgotPasswordResetPayload struct {
	UserID       string `json:"user_id"`
	Code         string `json:"code"`
	ExpireAt     int64  `json:"expire_at"`
	ExpireAtTime time.Time
	NewPassword  string `json:"new_password"`
}

func (payload ForgotPasswordResetPayload) Validate() error {
	if payload.UserID == "" {
		return skyerr.NewInvalidArgument("empty user_id", []string{"user_id"})
	}

	if payload.Code == "" {
		return skyerr.NewInvalidArgument("empty code", []string{"code"})
	}

	if payload.ExpireAt == 0 {
		return skyerr.NewInvalidArgument("empty expire_at", []string{"expire_at"})
	}

	if payload.NewPassword == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"new_password"})
	}

	return nil
}

// ForgotPasswordResetHandler reset user password with given code from email.
//
//  curl -X POST -H "Content-Type: application/json" \
//    -d @- http://localhost:3000/forgot_password/reset_password <<EOF
//  {
//    "user_id": "xxx",
//    "code": "xxx",
//    "expire_at": xxx, (utc timestamp)
//    "new_password": "xxx",
//  }
//  EOF
type ForgotPasswordResetHandler struct {
	CodeGenerator        *forgotpwdemail.CodeGenerator `dependency:"ForgotPasswordCodeGenerator"`
	PasswordChecker      dependency.PasswordChecker    `dependency:"PasswordChecker"`
	TokenStore           authtoken.Store               `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store                `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider             `dependency:"PasswordAuthProvider"`
	UserProfileStore     userprofile.Store             `dependency:"UserProfileStore"`
	TxContext            db.TxContext                  `dependency:"TxContext"`
	Logger               *logrus.Entry                 `dependency:"HandlerLogger"`
}

func (h ForgotPasswordResetHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h ForgotPasswordResetHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ForgotPasswordResetPayload{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	payload.ExpireAtTime = time.Unix(payload.ExpireAt, 0).UTC()

	return payload, nil
}

func (h ForgotPasswordResetHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(ForgotPasswordResetPayload)

	// check code expiration
	if timeNow().After(payload.ExpireAtTime) {
		h.Logger.Error("forgot password code expired")
		err = h.genericError()
		return
	}

	authInfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(payload.UserID, &authInfo); e != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(e).Error("user not found")
		err = h.genericError()
		return
	}

	// generate access-token
	token, err := h.TokenStore.NewToken(authInfo.ID)
	if err != nil {
		return
	}

	// Get Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID, token.AccessToken); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(err).Error("unable to get user profile")
		err = h.genericError()
		return
	}

	// Get password auth principals
	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(err).Error("unable to get password auth principals")
		err = h.genericError()
		return
	}

	hashedPassword := principals[0].HashedPassword
	expectedCode := h.CodeGenerator.Generate(authInfo, userProfile, hashedPassword, payload.ExpireAtTime)
	if payload.Code != expectedCode {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":       payload.UserID,
			"code":          payload.Code,
			"expected_code": expectedCode,
		}).Error("wrong forgot password reset password code")
		err = h.genericError()
		return
	}

	if err = h.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: payload.NewPassword,
	}); err != nil {
		return
	}

	// reset password
	for _, p := range principals {
		p.PlainPassword = payload.NewPassword
		err = h.PasswordAuthProvider.UpdatePrincipal(*p)
		if err != nil {
			return
		}
	}

	if err = h.TokenStore.Put(&token); err != nil {
		return
	}

	resp = "OK"
	return
}

func (h ForgotPasswordResetHandler) genericError() error {
	return skyerr.NewError(skyerr.ResourceNotFound, "user not found or code invalid")
}
