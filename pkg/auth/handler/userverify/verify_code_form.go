package userverify

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

// VerifyCodeFormHandlerFactory creates VerifyCodeFormHandler
type VerifyCodeFormHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new VerifyCodeFormHandler
func (f VerifyCodeFormHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &VerifyCodeFormHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h
}

type VerifyCodeFormPayload struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

const VerifyCodeFormSchema = `
{
	"$id": "#VerifyCodeForm",
	"type": "object",
	"properties": {
		"user_id": { "type": "string", "minLength": 1 },
		"code": { "type": "string", "minLength": 1 }
	},
	"required": ["user_id", "code"]
}
`

func decodeVerifyCodeFormRequest(request *http.Request) (payload VerifyCodeFormPayload, err error) {
	if err = request.ParseForm(); err != nil {
		return
	}

	payload = VerifyCodeFormPayload{
		UserID: request.Form.Get("user_id"),
		Code:   request.Form.Get("code"),
	}
	return
}

// VerifyCodeFormHandler reset user password with given code from email.
type VerifyCodeFormHandler struct {
	Validator                *validation.Validator          `dependency:"Validator"`
	AuthContext              coreAuth.ContextGetter         `dependency:"AuthContextGetter"`
	VerifyHTMLProvider       *userverify.VerifyHTMLProvider `dependency:"VerifyHTMLProvider"`
	UserVerificationProvider userverify.Provider            `dependency:"UserVerificationProvider"`
	AuthInfoStore            authinfo.Store                 `dependency:"AuthInfoStore"`
	PasswordAuthProvider     password.Provider              `dependency:"PasswordAuthProvider"`
	UserProfileStore         userprofile.Store              `dependency:"UserProfileStore"`
	HookProvider             hook.Provider                  `dependency:"HookProvider"`
	TxContext                db.TxContext                   `dependency:"TxContext"`
	Logger                   *logrus.Entry                  `dependency:"HandlerLogger"`
}

type resultTemplateContext struct {
	err        error
	payload    VerifyCodeFormPayload
	verifyCode userverify.VerifyCode
	user       model.User
}

func (h VerifyCodeFormHandler) prepareResultTemplateContext(r *http.Request, ctx *resultTemplateContext) (err error) {
	var payload VerifyCodeFormPayload
	payload, err = decodeVerifyCodeFormRequest(r)
	if err != nil {
		err = skyerr.NewBadRequest("invalid request form")
		return
	}

	if err = h.Validator.WithMessage("invalid request form").ValidateGoValue("#VerifyCodeForm", payload); err != nil {
		return
	}

	ctx.payload = payload

	authInfo := authinfo.AuthInfo{}
	if err = h.AuthInfoStore.GetAuth(payload.UserID, &authInfo); err != nil {
		return
	}

	userProfile, err := h.UserProfileStore.GetUserProfile(authInfo.ID)
	if err != nil {
		return
	}

	oldUser := model.NewUser(authInfo, userProfile)
	ctx.user = oldUser

	verifyCode, err := h.UserVerificationProvider.VerifyUser(h.PasswordAuthProvider, h.AuthInfoStore, &authInfo, payload.Code)
	if err != nil {
		return
	}

	user := model.NewUser(authInfo, userProfile)

	isVerified := authInfo.IsVerified()
	err = h.HookProvider.DispatchEvent(
		event.UserUpdateEvent{
			Reason:     event.UserUpdateReasonVerification,
			User:       oldUser,
			VerifyInfo: &authInfo.VerifyInfo,
			IsVerified: &isVerified,
		},
		&user,
	)

	ctx.user = user
	ctx.verifyCode = *verifyCode
	return
}

// HandleVerifyError handle the case when the given data (code, user_id) in the form is wrong
func (h VerifyCodeFormHandler) HandleVerifyError(rw http.ResponseWriter, templateCtx resultTemplateContext) {
	context := map[string]interface{}{
		"error": skyerr.AsAPIError(templateCtx.err),
	}

	if templateCtx.payload.Code != "" {
		context["code"] = templateCtx.payload.Code
	}

	if templateCtx.payload.UserID != "" {
		context["user_id"] = templateCtx.payload.UserID
	}

	url := h.VerifyHTMLProvider.ErrorRedirect(templateCtx.verifyCode.LoginIDKey, context)
	if url != nil {
		rw.Header().Set("Location", url.String())
		rw.WriteHeader(http.StatusFound)
		return
	}

	if templateCtx.user.ID != "" {
		context["user"] = templateCtx.user
	}

	html, htmlErr := h.VerifyHTMLProvider.ErrorHTML(templateCtx.verifyCode.LoginIDKey, context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusBadRequest)
	io.WriteString(rw, html)
}

func (h VerifyCodeFormHandler) HandleVerifySuccess(rw http.ResponseWriter, templateCtx resultTemplateContext) {
	context := map[string]interface{}{
		"code":    templateCtx.payload.Code,
		"user_id": templateCtx.payload.UserID,
	}

	url := h.VerifyHTMLProvider.SuccessRedirect(templateCtx.verifyCode.LoginIDKey, context)
	if url != nil {
		rw.Header().Set("Location", url.String())
		rw.WriteHeader(http.StatusFound)
		return
	}

	context["user"] = templateCtx.user

	html, htmlErr := h.VerifyHTMLProvider.SuccessHTML(templateCtx.verifyCode.LoginIDKey, context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, html)
}

func (h VerifyCodeFormHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	result, err := handler.Transactional(h.TxContext, func() (interface{}, error) {
		templateCtx := resultTemplateContext{}
		err := h.prepareResultTemplateContext(r, &templateCtx)
		return templateCtx, err
	})

	templateCtx := result.(resultTemplateContext)
	if err != nil {
		templateCtx.err = err
		h.HandleVerifyError(rw, templateCtx)
	} else {
		h.HandleVerifySuccess(rw, templateCtx)
	}
}
