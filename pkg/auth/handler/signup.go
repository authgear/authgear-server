package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func AttachSignupHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/signup", &SignupHandlerFactory{
		authDependency,
	}).Methods("POST")
	return server
}

type SignupHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f SignupHandlerFactory) NewHandler(ctx context.Context, tenantConfig config.TenantConfiguration) handler.Handler {
	h := &SignupHandler{}
	inject.DefaultInject(h, f.Dependency, ctx, tenantConfig)
	return handler.APIHandlerToHandler(h)
}

type SignupRequestPayload struct {
	AuthData         map[string]interface{} `json:"auth_data"`
	Password         string                 `json:"password"`
	Provider         string                 `json:"provider"`
	ProviderAuthData map[string]interface{} `json:"provider_auth_data"`
	// TODO:
	// RawProfile       map[string]interface{} `json:"profile"`
}

func (p SignupRequestPayload) Validate() error {
	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	return nil
}

func (p SignupRequestPayload) isAnonymous() bool {
	return len(p.AuthData) == 0 && p.Password == "" && p.Provider == ""
}

// SignupHandler handles signup request
type SignupHandler struct {
	AuthDataChecker      dependency.AuthDataChecker `dependency:"AuthDataChecker"`
	PasswordChecker      dependency.PasswordChecker `dependency:"PasswordChecker"`
	TokenStore           authtoken.Store            `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store             `dependency:"AuthInfoStore"`
	AuthPrincipalStore   principal.Store            `dependency:"AuthPrincipalStore"`
	PasswordAuthProvider password.Provider          `dependency:"PasswordAuthProvider"`
}

func (h SignupHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

func (h SignupHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SignupRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h SignupHandler) Handle(req interface{}, _ handler.AuthContext) (resp interface{}, err error) {
	payload := req.(SignupRequestPayload)

	if valid := h.AuthDataChecker.IsValid(payload.AuthData); !valid {
		err = skyerr.NewInvalidArgument("invalid auth data", []string{"auth_data"})
		return
	}

	// TODO: check duplicated keys in auth data and profile

	// validate password
	if err = h.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: payload.Password,
	}); err != nil {
		return
	}

	authContext := handler.AuthContext{}
	info := authinfo.NewAuthInfo()

	authContext.AuthInfo = &info

	// TODO: create user profile

	// Create AuthInfo
	if err = h.AuthInfoStore.CreateAuth(authContext.AuthInfo); err != nil {
		if err == skydb.ErrUserDuplicated {
			err = skyerr.NewError(skyerr.Duplicated, "user duplicated")
			return
		}

		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save auth info")
		return
	}

	principal := principal.New()

	if payload.isAnonymous() {
		panic("Unsupported signup anonymously")
		// info = authinfo.NewAnonymousAuthInfo()
	} else if payload.Provider != "" {
		panic("Unsupported signup with provider")
		// 	// Get AuthProvider and authenticates the user
		// 	logger.Debugf(`Client requested auth provider: "%v".`, p.Provider)
		// 	authProvider, err := h.ProviderRegistry.GetAuthProvider(p.Provider)
		// 	if err != nil {
		// 		response.Err = skyerr.NewInvalidArgument(err.Error(), []string{"provider"})
		// 		return
		// 	}
		// 	principalID, providerAuthData, err := authProvider.Login(payload.Context(), p.ProviderAuthData)
		// 	if err != nil {
		// 		response.Err = skyerr.NewError(skyerr.InvalidCredentials, "unable to login with the given credentials")
		// 		return
		// 	}
		// 	logger.Infof(`Client authenticated as principal: "%v" (provider: "%v").`, principalID, p.Provider)

		// 	// Create new user info and set updated auth data
		// 	info = skydb.NewProviderInfoAuthInfo(principalID, providerAuthData)
	} else {
		principal.Provider = "password"
		principal.UserID = info.ID
	}

	err = h.AuthPrincipalStore.CreatePrincipal(principal)
	if err != nil {
		return
	}

	err = h.createProviderEntry(payload, principal)
	if err != nil {
		return
	}

	tkn, err := h.TokenStore.NewToken(authContext.AuthInfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&tkn); err != nil {
		panic(err)
	}

	resp = response.NewAuthResponse(authContext, skydb.Record{}, tkn.AccessToken)

	// Populate the activity time to user
	now := timeNow()
	authContext.AuthInfo.LastSeenAt = &now
	// authContext.AuthInfo.IsPasswordSet = false
	if err = h.AuthInfoStore.UpdateAuth(authContext.AuthInfo); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	// TODO: Update user last login time
	// TODO: Audit

	return
}

func (h SignupHandler) createProviderEntry(payload SignupRequestPayload, prin principal.Principal) (err error) {
	if payload.isAnonymous() {
		return
	}

	if payload.Provider != "" {
		panic("Unsupported signup with provider")
	} else {
		err = h.PasswordAuthProvider.CreateEntry(prin.ID, payload.AuthData, payload.Password)
	}

	return
}
