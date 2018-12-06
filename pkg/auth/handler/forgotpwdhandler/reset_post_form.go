package forgotpwdhandler

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// ForgotPasswordResetPostFormHandlerFactory creates ForgotPasswordResetPostFormHandler
type ForgotPasswordResetPostFormHandlerFactory struct {
	Dependency auth.RequestDependencyMap
}

// NewHandler creates new ForgotPasswordResetPostFormHandler
func (f ForgotPasswordResetPostFormHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordResetPostFormHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f ForgotPasswordResetPostFormHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.Everybody{Allow: true}
}

// ForgotPasswordResetPostFormHandler reset user password with given code from email.
type ForgotPasswordResetPostFormHandler struct {
	CodeGenerator        *forgotpwdemail.CodeGenerator `dependency:"ForgotPasswordCodeGenerator"`
	PasswordChecker      dependency.PasswordChecker    `dependency:"PasswordChecker"`
	TokenStore           authtoken.Store               `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store                `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider             `dependency:"PasswordAuthProvider"`
	UserProfileStore     userprofile.Store             `dependency:"UserProfileStore"`
	TxContext            db.TxContext                  `dependency:"TxContext"`
	Logger               *logrus.Entry                 `dependency:"HandlerLogger"`
}

type ForgotPasswordResetPostFormHandlerResult struct {
	err         skyerr.Error
	payload     ForgotPasswordResetPayload
	userProfile userprofile.UserProfile
}

func (h ForgotPasswordResetPostFormHandler) WithTx() bool {
	return true
}

func (h ForgotPasswordResetPostFormHandler) DecodeRequest(request *http.Request) (payload ForgotPasswordResetPayload, err error) {
	return decodeForgotPasswordResetFormRequest(request)
}

func (h ForgotPasswordResetPostFormHandler) RenderRequestError(rw http.ResponseWriter, err skyerr.Error) {
	fmt.Fprintf(rw, "RenderRequestError: %+v", err)
}

func (h ForgotPasswordResetPostFormHandler) RenderErrorHTML(rw http.ResponseWriter, result ForgotPasswordResetPostFormHandlerResult) {
	fmt.Fprintf(rw, "RenderErrorHTML: %+v", result)
}

func (h ForgotPasswordResetPostFormHandler) RenderSuccessHTML(rw http.ResponseWriter, result ForgotPasswordResetPostFormHandlerResult) {
	fmt.Fprintf(rw, "RenderSuccessHTML: %+v", result)
}

func (h ForgotPasswordResetPostFormHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	payload, err := h.DecodeRequest(r)
	if err != nil {
		h.RenderRequestError(rw, skyerr.MakeError(err))
		return
	}

	if err := payload.Validate(); err != nil {
		h.RenderRequestError(rw, skyerr.MakeError(err))
		return
	}

	if err := h.TxContext.BeginTx(); err != nil {
		h.RenderRequestError(rw, skyerr.MakeError(err))
		return
	}

	result := ForgotPasswordResetPostFormHandlerResult{
		payload: payload,
	}
	defer func() {
		if result.err != nil {
			h.RenderErrorHTML(rw, result)
		} else {
			h.RenderSuccessHTML(rw, result)
		}
	}()

	defer func() {
		if result.err != nil {
			h.TxContext.RollbackTx()
		} else {
			h.TxContext.CommitTx()
		}
	}()

	// check code expiration
	if timeNow().After(payload.ExpireAtTime) {
		h.Logger.Error("forgot password code expired")
		result.err = genericResetPasswordError()
		return
	}

	authInfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(payload.UserID, &authInfo); e != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(e).Error("user not found")
		result.err = genericResetPasswordError()
		return
	}

	// Get Profile
	if result.userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(err).Error("unable to get user profile")
		result.err = genericResetPasswordError()
		return
	}

	// Get password auth principals
	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(err).Error("unable to get password auth principals")
		result.err = genericResetPasswordError()
		return
	}

	hashedPassword := principals[0].HashedPassword
	expectedCode := h.CodeGenerator.Generate(authInfo, result.userProfile, hashedPassword, payload.ExpireAtTime)
	if payload.Code != expectedCode {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":       payload.UserID,
			"code":          payload.Code,
			"expected_code": expectedCode,
		}).Error("wrong forgot password reset password code")
		result.err = genericResetPasswordError()
		return
	}

	resetPwdCtx := password.ResetPasswordRequestContext{
		PasswordChecker:      h.PasswordChecker,
		PasswordAuthProvider: h.PasswordAuthProvider,
	}

	if err := resetPwdCtx.ExecuteWithPrincipals(payload.NewPassword, principals); err != nil {
		result.err = skyerr.MakeError(err)
		return
	}
}
