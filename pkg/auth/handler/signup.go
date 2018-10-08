package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
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
	AuthData   map[string]interface{} `json:"auth_data"`
	Password   string                 `json:"password"`
	RawProfile map[string]interface{} `json:"profile"`
}

func (p SignupRequestPayload) Validate() error {
	if p.isAnonymous() {
		//no validation logic for anonymous sign up
	} else {
		if len(p.AuthData) == 0 {
			return skyerr.NewInvalidArgument("empty auth data", []string{"auth_data"})
		}

		if duplicatedKeys := p.duplicatedKeysInAuthDataAndProfile(); len(duplicatedKeys) > 0 {
			return skyerr.NewInvalidArgument("duplicated keys found in auth data in profile", duplicatedKeys)
		}

		if p.Password == "" {
			return skyerr.NewInvalidArgument("empty password", []string{"password"})
		}
	}

	return nil
}

func (p SignupRequestPayload) duplicatedKeysInAuthDataAndProfile() []string {
	keys := []string{}

	for k := range p.AuthData {
		if _, found := p.RawProfile[k]; found {
			keys = append(keys, k)
		}
	}

	return keys
}

func (p SignupRequestPayload) isAnonymous() bool {
	return len(p.AuthData) == 0 && p.Password == ""
}

// SignupHandler handles signup request
type SignupHandler struct {
	AuthDataChecker      dependency.AuthDataChecker  `dependency:"AuthDataChecker"`
	PasswordChecker      dependency.PasswordChecker  `dependency:"PasswordChecker"`
	UserProfileStore     dependency.UserProfileStore `dependency:"UserProfileStore,optional"`
	TokenStore           authtoken.Store             `dependency:"TokenStore"`
	AuthInfoStore        authinfo.Store              `dependency:"AuthInfoStore"`
	RoleStore            role.Store                  `dependency:"RoleStore"`
	PasswordAuthProvider password.Provider           `dependency:"PasswordAuthProvider"`
	AuditTrail           *coreAudit.Trail            `dependency:"AuditTrail,optional"`
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

	if h.UserProfileStore != nil {
		if err = h.UserProfileStore.CreateUserProfile(payload.RawProfile); err != nil {
			// TODO:
			// return proper error
			err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
			return
		}
	}

	// Get default roles
	defaultRoles, err := h.RoleStore.GetDefaultRoles()
	if err != nil {
		err = skyerr.NewError(skyerr.InternalQueryInvalid, "unable to query default roles")
		return
	}

	// Assign default roles
	info.Roles = defaultRoles

	authContext.AuthInfo = &info

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

	if h.AuditTrail != nil {
		h.AuditTrail.Log(coreAudit.Entry{
			AuthID: info.ID,
			Event:  coreAudit.EventSignup,
		})
	}

	return
}
