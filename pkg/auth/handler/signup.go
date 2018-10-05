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

func (f SignupHandlerFactory) NewHandler(request *http.Request) handler.Handler {
	h := &SignupHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h)
}

func (f SignupHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type SignupRequestPayload struct {
	AuthData         map[string]interface{} `json:"auth_data"`
	Password         string                 `json:"password"`
	Provider         string                 `json:"provider"`
	ProviderAuthData map[string]interface{} `json:"provider_auth_data"`
	RawProfile       map[string]interface{} `json:"profile"`
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
	AuthDataChecker      dependency.AuthDataChecker  `dependency:"AuthDataChecker"`
	PasswordChecker      dependency.PasswordChecker  `dependency:"PasswordChecker"`
	UserProfileStore     dependency.UserProfileStore `dependency:"UserProfileStore,optional"`
	TokenStore           authtoken.Store             `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store              `dependency:"AuthInfoStore"`
	PasswordAuthProvider password.Provider           `dependency:"PasswordAuthProvider"`
}

func (h SignupHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SignupRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h SignupHandler) Handle(req interface{}, _ context.AuthContext) (resp interface{}, err error) {
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

	authContext := context.AuthContext{}

	now := timeNow()
	info := authinfo.NewAuthInfo()
	info.LastLoginAt = &now

	authContext.AuthInfo = &info

	if h.UserProfileStore != nil {
		if err = h.UserProfileStore.CreateUserProfile(payload.RawProfile); err != nil {
			// TODO:
			// return proper error
			err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
			return
		}
	}

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

	// Create Principal
	principal := password.NewPrincipal()

	if payload.isAnonymous() {
		panic("Unsupported signup anonymously")
	} else if payload.Provider != "" {
		panic("Unsupported signup with provider")
	} else {
		principal.UserID = info.ID
		principal.AuthData = payload.AuthData
		principal.PlainPassword = payload.Password
	}

	err = h.PasswordAuthProvider.CreatePrincipal(principal)
	if err != nil {
		return
	}

	// Create auth token
	tkn, err := h.TokenStore.NewToken(authContext.AuthInfo.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&tkn); err != nil {
		panic(err)
	}

	resp = response.NewAuthResponse(authContext, skydb.Record{}, tkn.AccessToken)

	// Populate the activity time to user
	authContext.AuthInfo.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(authContext.AuthInfo); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	// TODO: Audit

	return
}
