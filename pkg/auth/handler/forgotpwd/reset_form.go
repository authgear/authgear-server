package forgotpwd

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// ForgotPasswordResetFormHandlerFactory creates ForgotPasswordResetFormHandler
type ForgotPasswordResetFormHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new ForgotPasswordResetFormHandler
func (f ForgotPasswordResetFormHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &ForgotPasswordResetFormHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h
}

type ForgotPasswordResetFormPayload struct {
	UserID          string
	Code            string
	ExpireAt        int64
	ExpireAtTime    time.Time
	NewPassword     string
	ConfirmPassword string
}

func decodeForgotPasswordResetFormRequest(request *http.Request) (payload ForgotPasswordResetFormPayload, err error) {
	if err = request.ParseForm(); err != nil {
		return
	}

	p := ForgotPasswordResetFormPayload{}
	p.UserID = request.Form.Get("user_id")
	p.Code = request.Form.Get("code")
	p.NewPassword = request.Form.Get("password")
	p.ConfirmPassword = request.Form.Get("confirm")

	expireAtStr := request.Form.Get("expire_at")
	var expireAt int64
	if expireAtStr != "" {
		if expireAt, err = strconv.ParseInt(expireAtStr, 10, 64); err != nil {
			return
		}
	}

	p.ExpireAt = expireAt
	p.ExpireAtTime = time.Unix(expireAt, 0).UTC()

	payload = p
	return
}

func (payload *ForgotPasswordResetFormPayload) Validate() error {
	if payload.UserID == "" {
		return skyerr.NewInvalidArgument("empty user_id", []string{"user_id"})
	}

	if payload.Code == "" {
		return skyerr.NewInvalidArgument("empty code", []string{"code"})
	}

	if payload.ExpireAt == 0 {
		return skyerr.NewInvalidArgument("empty expire_at", []string{"expire_at"})
	}

	return nil
}

// ForgotPasswordResetFormHandler reset user password with given code from email.
type ForgotPasswordResetFormHandler struct {
	CodeGenerator             *forgotpwdemail.CodeGenerator             `dependency:"ForgotPasswordCodeGenerator"`
	PasswordChecker           *audit.PasswordChecker                    `dependency:"PasswordChecker"`
	AuthInfoStore             authinfo.Store                            `dependency:"AuthInfoStore"`
	PasswordAuthProvider      password.Provider                         `dependency:"PasswordAuthProvider"`
	UserProfileStore          userprofile.Store                         `dependency:"UserProfileStore"`
	HookProvider              hook.Provider                             `dependency:"HookProvider"`
	ResetPasswordHTMLProvider *forgotpwdemail.ResetPasswordHTMLProvider `dependency:"ResetPasswordHTMLProvider"`
	TxContext                 db.TxContext                              `dependency:"TxContext"`
	Logger                    *logrus.Entry                             `dependency:"HandlerLogger"`
	TaskQueue                 async.Queue                               `dependency:"AsyncTaskQueue"`
}

type resultTemplateContext struct {
	err     skyerr.Error
	payload ForgotPasswordResetFormPayload
	user    model.User
}

func (h ForgotPasswordResetFormHandler) prepareResultTemplateContext(r *http.Request) (ctx resultTemplateContext, err error) {
	var payload ForgotPasswordResetFormPayload
	payload, err = decodeForgotPasswordResetFormRequest(r)
	if err != nil {
		return
	}

	if err = payload.Validate(); err != nil {
		return
	}

	ctx.payload = payload

	authInfo := authinfo.AuthInfo{}
	err = h.AuthInfoStore.GetAuth(payload.UserID, &authInfo)
	if err != nil {
		return
	}

	// Get Profile
	userProfile, err := h.UserProfileStore.GetUserProfile(payload.UserID)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": payload.UserID,
		}).WithError(err).Error("unable to get user profile")
		err = genericResetPasswordError()
		return
	}

	user := model.NewUser(authInfo, userProfile)
	ctx.user = user

	return
}

// HandleRequestError handle the case when the given data in the form is wrong, e.g. code, user_id, expire_at
func (h ForgotPasswordResetFormHandler) HandleRequestError(rw http.ResponseWriter, err skyerr.Error) {
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

// HandleResetError handle the case when the user input data in the form is wrong, e.g. password, confirm
func (h ForgotPasswordResetFormHandler) HandleResetError(rw http.ResponseWriter, templateCtx resultTemplateContext) {
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

	context["user"] = templateCtx.user

	// render the form again for failed post request
	html, htmlErr := h.ResetPasswordHTMLProvider.FormHTML(context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, html)
}

func (h ForgotPasswordResetFormHandler) HandleGetForm(rw http.ResponseWriter, templateCtx resultTemplateContext) {
	context := map[string]interface{}{
		"code":      templateCtx.payload.Code,
		"user_id":   templateCtx.payload.UserID,
		"user":      templateCtx.user,
		"expire_at": strconv.FormatInt(templateCtx.payload.ExpireAt, 10),
	}

	html, htmlErr := h.ResetPasswordHTMLProvider.FormHTML(context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, html)
}

func (h ForgotPasswordResetFormHandler) HandleResetSuccess(rw http.ResponseWriter, templateCtx resultTemplateContext) {
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

	context["user"] = templateCtx.user

	html, htmlErr := h.ResetPasswordHTMLProvider.SuccessHTML(context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, html)
}

func (h ForgotPasswordResetFormHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := handler.Transactional(h.TxContext, func() (_ interface{}, err error) {
		err = h.Handle(w, r)
		if err == nil {
			err = h.HookProvider.WillCommitTx()
		}
		return
	})
	if err == nil {
		h.HookProvider.DidCommitTx()
	}
}

func (h ForgotPasswordResetFormHandler) Handle(w http.ResponseWriter, r *http.Request) (err error) {
	var templateCtx resultTemplateContext
	if templateCtx, err = h.prepareResultTemplateContext(r); err != nil {
		h.HandleRequestError(w, skyerr.MakeError(err))
		return
	}

	// check code expiration
	if timeNow().After(templateCtx.payload.ExpireAtTime) {
		h.Logger.Error("forgot password code expired")
		h.HandleRequestError(w, genericResetPasswordError())
		return
	}

	authInfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(templateCtx.payload.UserID, &authInfo); e != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": templateCtx.payload.UserID,
		}).WithError(e).Error("user not found")
		h.HandleRequestError(w, genericResetPasswordError())
		return
	}

	// Get password auth principals
	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": templateCtx.payload.UserID,
		}).WithError(err).Error("unable to get password auth principals")
		h.HandleRequestError(w, genericResetPasswordError())
		return
	}

	hashedPassword := principals[0].HashedPassword

	expectedCode := h.CodeGenerator.Generate(authInfo, hashedPassword, templateCtx.payload.ExpireAtTime)
	if templateCtx.payload.Code != expectedCode {
		h.Logger.WithFields(map[string]interface{}{
			"user_id":       templateCtx.payload.UserID,
			"code":          templateCtx.payload.Code,
			"expected_code": expectedCode,
		}).Error("wrong forgot password reset password code")
		h.HandleRequestError(w, genericResetPasswordError())
		return
	}

	if r.Method == http.MethodGet {
		h.HandleGetForm(w, templateCtx)
		return
	}

	h.resetPassword(w, templateCtx, authInfo, principals)

	// password house keeper
	h.TaskQueue.Enqueue(task.PwHousekeeperTaskName, task.PwHousekeeperTaskParam{
		AuthID: authInfo.ID,
	}, nil)

	return templateCtx.err
}

func (h ForgotPasswordResetFormHandler) resetPassword(
	rw http.ResponseWriter,
	templateCtx resultTemplateContext,
	authInfo authinfo.AuthInfo,
	principals []*password.Principal,
) {
	var err error
	defer func() {
		if err != nil {
			templateCtx.err = skyerr.MakeError(err)
			h.HandleResetError(rw, templateCtx)
		} else {
			h.HandleResetSuccess(rw, templateCtx)
		}
	}()

	resetPwdCtx := password.ResetPasswordRequestContext{
		PasswordChecker:      h.PasswordChecker,
		PasswordAuthProvider: h.PasswordAuthProvider,
	}

	if templateCtx.payload.NewPassword == "" {
		err = skyerr.NewInvalidArgument("empty password", []string{"password"})
		return
	}

	if templateCtx.payload.NewPassword != templateCtx.payload.ConfirmPassword {
		err = skyerr.NewInvalidArgument("confirm password does not match the password", []string{"password", "confirm"})
		return
	}

	err = resetPwdCtx.ExecuteWithPrincipals(templateCtx.payload.NewPassword, principals)
	if err != nil {
		return
	}

	var profile userprofile.UserProfile
	if profile, err = h.UserProfileStore.GetUserProfile(authInfo.ID); err != nil {
		return
	}

	user := model.NewUser(authInfo, profile)

	err = h.HookProvider.DispatchEvent(
		event.PasswordUpdateEvent{
			Reason: event.PasswordUpdateReasonResetPassword,
			User:   user,
		},
		&user,
	)
}
