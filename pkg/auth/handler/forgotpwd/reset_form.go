package forgotpwd

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreaudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
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
	UserID          string `json:"user_id"`
	Code            string `json:"code"`
	ExpireAt        int64  `json:"expire_at"`
	ExpireAtTime    time.Time
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// nolint: gosec
const ForgotPasswordResetPageSchema = `
{
	"$id": "#ForgotPasswordResetPage",
	"type": "object",
	"properties": {
		"user_id": { "type": "string", "minLength": 1 },
		"code": { "type": "string", "minLength": 1 },
		"expire_at": { "type": "integer", "minimum": 1 }
	},
	"required": ["user_id", "code", "expire_at"]
}
`

// nolint: gosec
const ForgotPasswordResetFormSchema = `
{
	"$id": "#ForgotPasswordResetForm",
	"type": "object",
	"properties": {
		"user_id": { "type": "string", "minLength": 1 },
		"code": { "type": "string", "minLength": 1 },
		"expire_at": { "type": "integer", "minimum": 1 },
		"new_password": { "type": "string", "minLength": 1 },
		"confirm_password": { "type": "string", "minLength": 1 }
	},
	"required": ["user_id", "code", "expire_at", "new_password", "confirm_password"]
}
`

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

// ForgotPasswordResetFormHandler reset user password with given code from email.
type ForgotPasswordResetFormHandler struct {
	Validator                 *validation.Validator                     `dependency:"Validator"`
	CodeGenerator             *forgotpwdemail.CodeGenerator             `dependency:"ForgotPasswordCodeGenerator"`
	PasswordChecker           *audit.PasswordChecker                    `dependency:"PasswordChecker"`
	AuthInfoStore             authinfo.Store                            `dependency:"AuthInfoStore"`
	PasswordAuthProvider      password.Provider                         `dependency:"PasswordAuthProvider"`
	UserProfileStore          userprofile.Store                         `dependency:"UserProfileStore"`
	HookProvider              hook.Provider                             `dependency:"HookProvider"`
	ResetPasswordHTMLProvider *forgotpwdemail.ResetPasswordHTMLProvider `dependency:"ResetPasswordHTMLProvider"`
	TxContext                 db.TxContext                              `dependency:"TxContext"`
	Logger                    *logrus.Entry                             `dependency:"HandlerLogger"`
	AuditTrail                coreaudit.Trail                           `dependency:"AuditTrail"`
	TimeProvider              coreTime.Provider                         `dependency:"TimeProvider"`
	TaskQueue                 async.Queue                               `dependency:"AsyncTaskQueue"`
}

type resultTemplateContext struct {
	err     error
	payload ForgotPasswordResetFormPayload
	user    model.User
}

func (h ForgotPasswordResetFormHandler) prepareResultTemplateContext(r *http.Request) (ctx resultTemplateContext, err error) {
	var payload ForgotPasswordResetFormPayload
	payload, err = decodeForgotPasswordResetFormRequest(r)
	if err != nil {
		err = skyerr.NewBadRequest("invalid request form")
		return
	}

	var schema string
	if r.Method == http.MethodGet {
		schema = "#ForgotPasswordResetPage"
	} else {
		schema = "#ForgotPasswordResetForm"
	}
	if err = h.Validator.WithMessage("invalid request form").ValidateGoValue(schema, payload); err != nil {
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
		return
	}

	user := model.NewUser(authInfo, userProfile)
	ctx.user = user

	return
}

// HandleRequestError handle the case when the given data in the form is wrong, e.g. code, user_id, expire_at
func (h ForgotPasswordResetFormHandler) HandleRequestError(rw http.ResponseWriter, err error) {
	context := map[string]interface{}{
		"error": skyerr.AsAPIError(err),
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
		"error":     skyerr.AsAPIError(templateCtx.err),
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
		h.HandleRequestError(w, err)
		return
	}

	if r.Method == http.MethodGet {
		h.HandleGetForm(w, templateCtx)
		return
	}

	err = h.resetPassword(&templateCtx.user, templateCtx.payload)
	if err != nil {
		templateCtx.err = err
		h.HandleResetError(w, templateCtx)
	} else {
		h.HandleResetSuccess(w, templateCtx)
	}

	return templateCtx.err
}

func (h ForgotPasswordResetFormHandler) resetPassword(
	user *model.User,
	payload ForgotPasswordResetFormPayload,
) error {
	if payload.NewPassword != payload.ConfirmPassword {
		return NewPasswordResetFailed(PasswordNotMatched, "confirm password does not match new password")
	}

	now := h.TimeProvider.NowUTC()
	err := passwordReseter{
		CodeGenerator:        h.CodeGenerator,
		PasswordChecker:      h.PasswordChecker,
		AuthInfoStore:        h.AuthInfoStore,
		PasswordAuthProvider: h.PasswordAuthProvider,
	}.resetPassword(
		payload.UserID,
		now,
		payload.ExpireAtTime,
		payload.Code,
		payload.NewPassword,
	)
	if err != nil {
		return err
	}

	err = h.HookProvider.DispatchEvent(
		event.PasswordUpdateEvent{
			Reason: event.PasswordUpdateReasonResetPassword,
			User:   *user,
		},
		user,
	)
	if err != nil {
		return err
	}

	h.AuditTrail.Log(coreaudit.Entry{
		UserID: user.ID,
		Event:  coreaudit.EventResetPassword,
		Data: map[string]interface{}{
			"type": "forgot_password",
		},
	})

	// password house keeper
	h.TaskQueue.Enqueue(task.PwHousekeeperTaskName, task.PwHousekeeperTaskParam{
		AuthID: user.ID,
	}, nil)

	return nil
}
