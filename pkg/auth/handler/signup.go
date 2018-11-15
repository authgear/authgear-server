package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/db"
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
	}).Methods("OPTIONS", "POST")
	return server
}

type SignupHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f SignupHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &SignupHandler{}
	inject.DefaultInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(h, h.TxContext)
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

func (p SignupRequestPayload) mergedProfile() map[string]interface{} {
	// Assume duplicatedKeysInAuthDataAndProfile is called before this
	profile := make(map[string]interface{})
	for k := range p.AuthData {
		profile[k] = p.AuthData[k]
	}
	for k := range p.RawProfile {
		profile[k] = p.RawProfile[k]
	}
	return profile
}

// SignupHandler handles signup request
type SignupHandler struct {
	AuthDataChecker       dependency.AuthDataChecker `dependency:"AuthDataChecker"`
	PasswordChecker       dependency.PasswordChecker `dependency:"PasswordChecker"`
	UserProfileStore      userprofile.Store          `dependency:"UserProfileStore"`
	TokenStore            authtoken.Store            `dependency:"TokenStore"`
	AuthInfoStore         authinfo.Store             `dependency:"AuthInfoStore"`
	RoleStore             role.Store                 `dependency:"RoleStore"`
	PasswordAuthProvider  password.Provider          `dependency:"PasswordAuthProvider"`
	AnonymousAuthProvider anonymous.Provider         `dependency:"AnonymousAuthProvider"`
	AuditTrail            coreAudit.Trail            `dependency:"AuditTrail"`
	WelcomeEmailSendTask  *welcemail.SendTask        `dependency:"WelcomeEmailSendTask,optional"`
	TxContext             db.TxContext               `dependency:"TxContext"`
	Logger                *logrus.Entry              `dependency:"HandlerLogger"`
}

func (h SignupHandler) WithTx() bool {
	return true
}

func (h SignupHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SignupRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)
	return payload, err
}

func (h SignupHandler) Handle(req interface{}) (resp interface{}, err error) {
	payload := req.(SignupRequestPayload)

	err = h.verifyPayload(payload)
	if err != nil {
		return
	}

	now := timeNow()
	info := authinfo.NewAuthInfo()
	info.LastLoginAt = &now

	// Get default roles
	defaultRoles, err := h.RoleStore.GetDefaultRoles()
	if err != nil {
		err = skyerr.NewError(skyerr.InternalQueryInvalid, "unable to query default roles")
		return
	}

	// Assign default roles
	info.Roles = defaultRoles

	// Create AuthInfo
	if err = h.AuthInfoStore.CreateAuth(&info); err != nil {
		if err == skydb.ErrUserDuplicated {
			err = skyerr.NewError(skyerr.Duplicated, "user duplicated")
			return
		}

		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save auth info")
		return
	}

	// Create Profile
	var userProfile userprofile.UserProfile
	if userProfile, err = h.UserProfileStore.CreateUserProfile(info.ID, &info, payload.mergedProfile()); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
		return
	}

	// Create Principal
	err = h.createPrincipal(payload, info)
	if err != nil {
		return
	}

	// Create auth token
	tkn, err := h.TokenStore.NewToken(info.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&tkn); err != nil {
		panic(err)
	}

	resp = response.NewAuthResponse(info, userProfile, tkn.AccessToken)

	// Populate the activity time to user
	info.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	h.AuditTrail.Log(coreAudit.Entry{
		AuthID: info.ID,
		Event:  coreAudit.EventSignup,
	})

	if h.WelcomeEmailSendTask != nil {
		if email, ok := userProfile.Data["email"].(string); ok {
			h.WelcomeEmailSendTask.Request <- welcemail.SendTaskRequest{
				Email:       email,
				UserProfile: userProfile,
				Logger:      h.Logger,
			}
		}
	}

	return
}

func (h SignupHandler) verifyPayload(payload SignupRequestPayload) (err error) {
	if payload.isAnonymous() {
		return
	}

	if valid := h.AuthDataChecker.IsValid(payload.AuthData); !valid {
		err = skyerr.NewInvalidArgument("invalid auth data", []string{"auth_data"})
		return
	}

	// validate password
	err = h.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: payload.Password,
	})

	return
}

func (h SignupHandler) createPrincipal(payload SignupRequestPayload, info authinfo.AuthInfo) (err error) {
	if !payload.isAnonymous() {
		principal := password.NewPrincipal()
		principal.UserID = info.ID
		principal.AuthData = payload.AuthData
		principal.PlainPassword = payload.Password

		err = h.PasswordAuthProvider.CreatePrincipal(principal)
	} else {
		principal := anonymous.NewPrincipal()
		principal.UserID = info.ID

		err = h.AnonymousAuthProvider.CreatePrincipal(principal)
	}

	return
}
