package userverify

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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

// ProvideAuthzPolicy provides authorization policy of handler
func (f VerifyCodeFormHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.Everybody{Allow: true}
}

type VerifyCodeFormPayload struct {
	UserID string
	Code   string
}

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

func (payload *VerifyCodeFormPayload) Validate() error {
	if payload.UserID == "" {
		return skyerr.NewInvalidArgument("empty user_id", []string{"user_id"})
	}

	if payload.Code == "" {
		return skyerr.NewInvalidArgument("empty code", []string{"code"})
	}

	return nil
}

// VerifyCodeFormHandler reset user password with given code from email.
type VerifyCodeFormHandler struct {
	VerifyHTMLProvider     *userverify.VerifyHTMLProvider    `dependency:"VerifyHTMLProvider"`
	VerifyCodeStore        userverify.Store                  `dependency:"VerifyCodeStore"`
	UserProfileStore       userprofile.Store                 `dependency:"UserProfileStore"`
	AuthInfoStore          authinfo.Store                    `dependency:"AuthInfoStore"`
	PasswordAuthProvider   password.Provider                 `dependency:"PasswordAuthProvider"`
	UpdateVerifiedFlagFunc userverify.UpdateVerifiedFlagFunc `dependency:"UpdateVerifiedFlagFunc"`
	TxContext              db.TxContext                      `dependency:"TxContext"`
	Logger                 *logrus.Entry                     `dependency:"HandlerLogger"`
}

type resultTemplateContext struct {
	err         skyerr.Error
	payload     VerifyCodeFormPayload
	verifyCode  userverify.VerifyCode
	userProfile userprofile.UserProfile
}

func (h VerifyCodeFormHandler) prepareResultTemplateContext(r *http.Request, ctx *resultTemplateContext) (err error) {
	var payload VerifyCodeFormPayload
	payload, err = decodeVerifyCodeFormRequest(r)
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
		return
	}

	// Get verify code
	verifyCodeReq := getAndValidateCodeRequest{
		VerifyCodeStore: h.VerifyCodeStore,
		Logger:          h.Logger,
	}

	if ctx.verifyCode, err = verifyCodeReq.execute(payload.Code, ctx.userProfile); err != nil {
		return
	}

	return
}

// HandleVerifyError handle the case when the given data (code, user_id) in the form is wrong
func (h VerifyCodeFormHandler) HandleVerifyError(rw http.ResponseWriter, templateCtx resultTemplateContext) {
	context := map[string]interface{}{
		"error": templateCtx.err.Message(),
	}

	if templateCtx.payload.Code != "" {
		context["code"] = templateCtx.payload.Code
	}

	if templateCtx.payload.UserID != "" {
		context["user_id"] = templateCtx.payload.UserID
	}

	url := h.VerifyHTMLProvider.ErrorRedirect(templateCtx.verifyCode.RecordKey, context)
	if url != nil {
		rw.Header().Set("Location", url.String())
		rw.WriteHeader(http.StatusFound)
		return
	}

	if templateCtx.userProfile.ID != "" {
		context["user"] = templateCtx.userProfile
	}

	html, htmlErr := h.VerifyHTMLProvider.ErrorHTML(templateCtx.verifyCode.RecordKey, context)
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

	url := h.VerifyHTMLProvider.SuccessRedirect(templateCtx.verifyCode.RecordKey, context)
	if url != nil {
		rw.Header().Set("Location", url.String())
		rw.WriteHeader(http.StatusFound)
		return
	}

	context["user"] = templateCtx.userProfile

	html, htmlErr := h.VerifyHTMLProvider.SuccessHTML(templateCtx.verifyCode.RecordKey, context)
	if htmlErr != nil {
		panic(htmlErr)
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, html)
}

func (h VerifyCodeFormHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	templateCtx := resultTemplateContext{}

	var err error
	defer func() {
		// result handling
		if err != nil {
			templateCtx.err = skyerr.MakeError(err)
			h.HandleVerifyError(rw, templateCtx)
		} else {
			h.HandleVerifySuccess(rw, templateCtx)
		}
	}()

	if err = h.TxContext.BeginTx(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			h.TxContext.RollbackTx()
		} else {
			h.TxContext.CommitTx()
		}
	}()

	if err = h.prepareResultTemplateContext(r, &templateCtx); err != nil {
		return
	}

	// Update code
	templateCtx.verifyCode.Consumed = true
	if err = h.VerifyCodeStore.UpdateVerifyCode(&templateCtx.verifyCode); err != nil {
		return
	}

	// Update user
	authInfo := &authinfo.AuthInfo{}
	if err = h.AuthInfoStore.GetAuth(templateCtx.payload.UserID, authInfo); err != nil {
		return
	}

	principals, err := h.PasswordAuthProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		return
	}

	authInfo.VerifyInfo[templateCtx.verifyCode.RecordValue] = true
	h.UpdateVerifiedFlagFunc(authInfo, principals)

	if err = h.AuthInfoStore.UpdateAuth(authInfo); err != nil {
		return
	}

	return
}
