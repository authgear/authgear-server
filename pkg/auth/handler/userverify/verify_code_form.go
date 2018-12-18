package userverify

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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
	UserProfileStore userprofile.Store `dependency:"UserProfileStore"`
	TxContext        db.TxContext      `dependency:"TxContext"`
	Logger           *logrus.Entry     `dependency:"HandlerLogger"`
}

type resultTemplateContext struct {
	err         skyerr.Error
	payload     VerifyCodeFormPayload
	userProfile userprofile.UserProfile
}

func (h VerifyCodeFormHandler) prepareResultTemplateContext(r *http.Request) (ctx resultTemplateContext, err error) {
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

	return
}

func (h VerifyCodeFormHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := h.TxContext.BeginTx(); err != nil {
		// handle error
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
		// handle error
		return
	}

	fmt.Fprintf(rw, "%+v", templateCtx)

	// Handle logic
}
