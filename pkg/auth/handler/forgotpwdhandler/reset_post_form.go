package forgotpwdhandler

import (
	"io"
	"net/http"
	"strconv"

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
	CodeGenerator             *forgotpwdemail.CodeGenerator             `dependency:"ForgotPasswordCodeGenerator"`
	PasswordChecker           dependency.PasswordChecker                `dependency:"PasswordChecker"`
	TokenStore                authtoken.Store                           `dependency:"TokenStore"`
	AuthInfoStore             authinfo.Store                            `dependency:"AuthInfoStore"`
	PasswordAuthProvider      password.Provider                         `dependency:"PasswordAuthProvider"`
	UserProfileStore          userprofile.Store                         `dependency:"UserProfileStore"`
	ResetPasswordHTMLProvider *forgotpwdemail.ResetPasswordHTMLProvider `dependency:"ResetPasswordHTMLProvider"`
	TxContext                 db.TxContext                              `dependency:"TxContext"`
	Logger                    *logrus.Entry                             `dependency:"HandlerLogger"`
}

type resultTemplateContext struct {
	err         skyerr.Error
	payload     ForgotPasswordResetPayload
	userProfile userprofile.UserProfile
}

func (h ForgotPasswordResetPostFormHandler) WithTx() bool {
	return true
}

func (h ForgotPasswordResetPostFormHandler) prepareResultTemplateContext(r *http.Request) (ctx resultTemplateContext, err error) {
	var payload ForgotPasswordResetPayload
	payload, err = decodeForgotPasswordResetFormRequest(r)
	if err != nil {
		return
	}

	if err = payload.Validate(); err != nil {
		return
	}

	ctx.payload = payload

	// Get Profile
	if ctx.userProfile, err = h.UserProfileStore.GetUserProfile(payload.UserID); err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(err).Error("unable to get user profile")
		err = genericResetPasswordError()
		return
	}

	return
}

func (h ForgotPasswordResetPostFormHandler) RenderRequestError(rw http.ResponseWriter, err skyerr.Error) {
	context := map[string]interface{}{
		"error": err.Message(),
	}

	url := h.ResetPasswordHTMLProvider.ErrorRedirect(context)
	if url != nil {
		rw.Header().Set("Location", url.String())
		rw.WriteHeader(http.StatusFound)
		return
	}

	html, htmlErr := h.ResetPasswordHTMLProvider.ErrorHTML(context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusBadRequest)
	io.WriteString(rw, html)
}

func (h ForgotPasswordResetPostFormHandler) RenderErrorHTML(rw http.ResponseWriter, templateCtx resultTemplateContext) {
	context := map[string]interface{}{
		"error":     templateCtx.err.Message(),
		"code":      templateCtx.payload.Code,
		"user_id":   templateCtx.payload.UserID,
		"expire_at": strconv.FormatInt(templateCtx.payload.ExpireAt, 10),
	}

	url := h.ResetPasswordHTMLProvider.ErrorRedirect(context)
	if url != nil {
		rw.Header().Set("Location", url.String())
		rw.WriteHeader(http.StatusFound)
		return
	}

	context["user"] = templateCtx.userProfile.ToMap()

	// render the form again for failed post request
	html, htmlErr := h.ResetPasswordHTMLProvider.FormHTML(context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, html)
}

func (h ForgotPasswordResetPostFormHandler) RenderSuccessHTML(rw http.ResponseWriter, templateCtx resultTemplateContext) {
	context := map[string]interface{}{
		"code":      templateCtx.payload.Code,
		"user_id":   templateCtx.payload.UserID,
		"expire_at": strconv.FormatInt(templateCtx.payload.ExpireAt, 10),
	}

	url := h.ResetPasswordHTMLProvider.SuccessRedirect(context)
	if url != nil {
		rw.WriteHeader(http.StatusFound)
		rw.Header().Set("Location", url.String())
		return
	}

	context["user"] = templateCtx.userProfile.ToMap()

	html, htmlErr := h.ResetPasswordHTMLProvider.SuccessHTML(context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, html)
}

func (h ForgotPasswordResetPostFormHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := h.TxContext.BeginTx(); err != nil {
		h.RenderRequestError(rw, skyerr.MakeError(err))
		return
	}

	var err error
	defer func() {
		if err != nil {
			h.TxContext.RollbackTx()
		} else {
			h.TxContext.CommitTx()
		}
	}()

	var templateCtx resultTemplateContext
	if templateCtx, err = h.prepareResultTemplateContext(r); err != nil {
		h.RenderRequestError(rw, skyerr.MakeError(err))
		return
	}

	// check code expiration
	if timeNow().After(templateCtx.payload.ExpireAtTime) {
		h.Logger.Error("forgot password code expired")
		h.RenderRequestError(rw, genericResetPasswordError())
		return
	}

	authInfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(templateCtx.payload.UserID, &authInfo); e != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": templateCtx.payload.UserID,
		}).WithError(e).Error("user not found")
		h.RenderRequestError(rw, genericResetPasswordError())
		return
	}

	// Get password auth principals
	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": templateCtx.payload.UserID,
		}).WithError(err).Error("unable to get password auth principals")
		h.RenderRequestError(rw, genericResetPasswordError())
		return
	}

	hashedPassword := principals[0].HashedPassword
	expectedCode := h.CodeGenerator.Generate(authInfo, templateCtx.userProfile, hashedPassword, templateCtx.payload.ExpireAtTime)
	if templateCtx.payload.Code != expectedCode {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":       templateCtx.payload.UserID,
			"code":          templateCtx.payload.Code,
			"expected_code": expectedCode,
		}).Error("wrong forgot password reset password code")
		h.RenderRequestError(rw, genericResetPasswordError())
		return
	}

	h.resetPassword(rw, templateCtx, principals)
}

func (h ForgotPasswordResetPostFormHandler) resetPassword(rw http.ResponseWriter, templateCtx resultTemplateContext, principals []*password.Principal) {
	resetPwdCtx := password.ResetPasswordRequestContext{
		PasswordChecker:      h.PasswordChecker,
		PasswordAuthProvider: h.PasswordAuthProvider,
	}

	if err := resetPwdCtx.ExecuteWithPrincipals(templateCtx.payload.NewPassword, principals); err != nil {
		templateCtx.err = skyerr.MakeError(err)
		h.RenderErrorHTML(rw, templateCtx)
	} else {
		h.RenderSuccessHTML(rw, templateCtx)
	}
}
