package userverify

import (
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
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
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
	}).Methods("OPTIONS", "POST", "GET")
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
	return handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext)
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

// @JSONSchema
const VerifyCodeRequestSchema = `
{
	"$id": "#VerifyCodeRequest",
	"type": "object",
	"properties": {
		"code": { "type": "string" }
	}
}
`

func (payload VerifyCodePayload) Validate() error {
	if payload.Code == "" {
		return skyerr.NewInvalidArgument("empty code", []string{"code"})
	}

	return nil
}

/*
	@Operation POST /verify_code - Submit verification code
		Verify user using received verification code.

		@Tag User Verification
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@RequestBody
			@JSONSchema {VerifyCodeRequest}

		@Response 200 {EmptyResponse}

		@Callback user_update {UserUpdateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type VerifyCodeHandler struct {
	TxContext                db.TxContext           `dependency:"TxContext"`
	AuthContext              coreAuth.ContextGetter `dependency:"AuthContextGetter"`
	UserVerificationProvider userverify.Provider    `dependency:"UserVerificationProvider"`
	AuthInfoStore            authinfo.Store         `dependency:"AuthInfoStore"`
	PasswordAuthProvider     password.Provider      `dependency:"PasswordAuthProvider"`
	UserProfileStore         userprofile.Store      `dependency:"UserProfileStore"`
	HookProvider             hook.Provider          `dependency:"HookProvider"`
	Logger                   *logrus.Entry          `dependency:"HandlerLogger"`
}

func (h VerifyCodeHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h VerifyCodeHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := VerifyCodePayload{}
	if err := handler.DecodeJSONBody(request, &payload); err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "fails to decode the request payload")
	}

	return payload, nil
}

func (h VerifyCodeHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(VerifyCodePayload)
	authInfo := h.AuthContext.AuthInfo()

	var userProfile userprofile.UserProfile
	userProfile, err = h.UserProfileStore.GetUserProfile(authInfo.ID)
	if err != nil {
		h.Logger.WithFields(map[string]interface{}{
			"user_id": authInfo.ID,
		}).WithError(err).Error("unable to get user profile")
		return
	}

	oldUser := model.NewUser(*authInfo, userProfile)

	_, err = h.UserVerificationProvider.VerifyUser(h.PasswordAuthProvider, h.AuthInfoStore, authInfo, payload.Code)
	if err != nil {
		return
	}

	user := model.NewUser(*authInfo, userProfile)

	err = h.HookProvider.DispatchEvent(
		event.UserUpdateEvent{
			Reason:     event.UserUpdateReasonVerification,
			User:       oldUser,
			VerifyInfo: &authInfo.VerifyInfo,
			IsVerified: &authInfo.Verified,
		},
		&user,
	)

	resp = map[string]string{}
	return
}
