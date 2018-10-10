package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/handler/context"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachResetPasswordHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/reset_password", &ResetPasswordHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type ResetPasswordHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f ResetPasswordHandlerFactory) NewHandler(request *http.Request) handler.Handler {
	h := &ResetPasswordHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h)
}

func (f ResetPasswordHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return policy.AllOf(
		authz.PolicyFunc(policy.DenyNoAccessKey),
		authz.PolicyFunc(policy.RequireAuthenticated),
		authz.PolicyFunc(policy.DenyDisabledUser),
	)
}

type ResetPasswordRequestPayload struct {
	AuthInfoID string `json:"auth_id"`
	Password   string `json:"password"`
}

func (p ResetPasswordRequestPayload) Validate() error {
	if p.AuthInfoID == "" {
		return skyerr.NewInvalidArgument("invalid auth id", []string{"auth_id"})
	}

	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

// ResetPasswordHandler handles signup request
type ResetPasswordHandler struct {
	PasswordChecker      dependency.PasswordChecker `dependency:"PasswordChecker"`
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
}

func (h ResetPasswordHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := ResetPasswordRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h ResetPasswordHandler) Handle(req interface{}, _ context.AuthContext) (resp interface{}, err error) {
	payload := req.(ResetPasswordRequestPayload)

	authinfo := authinfo.AuthInfo{}
	if e := h.AuthInfoStore.GetAuth(payload.AuthInfoID, &authinfo); e != nil {
		if err == skydb.ErrUserNotFound {
			// logger.Info("Auth info not found when setting disabled user status")
			err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
		// logger.WithError(err).Error("Unable to get auth info when setting disabled user status")
		err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
		return
	}

	if err = h.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: payload.Password,
	}); err != nil {
		return
	}

	// reset password
	principal := password.Principal{}
	err = h.PasswordAuthProvider.GetPrincipalByUserID(authinfo.ID, &principal)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
			return
		}
		return
	}
	principal.PlainPassword = payload.Password
	err = h.PasswordAuthProvider.UpdatePrincipal(principal)
	if err != nil {
		return
	}

	// generate access-token
	token, err := h.TokenStore.NewToken(authinfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	authContext := context.AuthContext{}
	authContext.AuthInfo = &authinfo

	resp = response.NewAuthResponse(authContext, skydb.Record{}, token.AccessToken)

	// TODO: Audit

	return
}
