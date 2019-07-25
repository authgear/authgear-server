package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/event"

	"github.com/skygeario/skygear-server/pkg/core/utils"

	"github.com/skygeario/skygear-server/pkg/auth/task"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrUserDuplicated = skyerr.NewError(skyerr.Duplicated, "user duplicated")

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
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext)
}

func (f SignupHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type SignupRequestPayload struct {
	LoginIDs []password.LoginID     `json:"login_ids"`
	Realm    string                 `json:"realm"`
	Password string                 `json:"password"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (p SignupRequestPayload) Validate() error {
	if len(p.LoginIDs) == 0 {
		return skyerr.NewInvalidArgument("empty login_ids", []string{"login_ids"})
	}

	if p.Password == "" {
		return skyerr.NewInvalidArgument("empty password", []string{"password"})
	}

	for _, loginID := range p.LoginIDs {
		if !loginID.IsValid() {
			return skyerr.NewInvalidArgument("invalid login_ids", []string{"login_ids"})
		}
	}

	if p.duplicatedLoginIDs() {
		return skyerr.NewInvalidArgument("duplicated login_ids", []string{"login_ids"})
	}

	return nil
}

func (p SignupRequestPayload) duplicatedLoginIDs() bool {
	loginIDs := []string{}

	for _, loginID := range p.LoginIDs {
		found := utils.StringSliceContains(loginIDs, loginID.Value)

		if found {
			return found
		}

		loginIDs = append(loginIDs, loginID.Value)
	}

	return false
}

// SignupHandler handles signup request
type SignupHandler struct {
	PasswordChecker         *authAudit.PasswordChecker                         `dependency:"PasswordChecker"`
	UserProfileStore        userprofile.Store                                  `dependency:"UserProfileStore"`
	TokenStore              authtoken.Store                                    `dependency:"TokenStore"`
	AuthInfoStore           authinfo.Store                                     `dependency:"AuthInfoStore"`
	PasswordAuthProvider    password.Provider                                  `dependency:"PasswordAuthProvider"`
	IdentityProvider        principal.IdentityProvider                         `dependency:"IdentityProvider"`
	AuditTrail              audit.Trail                                        `dependency:"AuditTrail"`
	WelcomeEmailEnabled     bool                                               `dependency:"WelcomeEmailEnabled"`
	WelcomeEmailDestination config.WelcomeEmailDestination                     `dependency:"WelcomeEmailDestination"`
	AutoSendUserVerifyCode  bool                                               `dependency:"AutoSendUserVerifyCodeOnSignup"`
	UserVerifyLoginIDKeys   map[string]config.UserVerificationKeyConfiguration `dependency:"UserVerifyLoginIDKeys"`
	TxContext               db.TxContext                                       `dependency:"TxContext"`
	Logger                  *logrus.Entry                                      `dependency:"HandlerLogger"`
	TaskQueue               async.Queue                                        `dependency:"AsyncTaskQueue"`
	HookProvider            hook.Provider                                      `dependency:"HookProvider"`
}

func (h SignupHandler) WithTx() bool {
	return true
}

func (h SignupHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := SignupRequestPayload{}
	err := json.NewDecoder(request.Body).Decode(&payload)

	if err != nil {
		return nil, err
	}

	// Avoid { metadata: null } in the response user object
	if payload.Metadata == nil {
		payload.Metadata = make(map[string]interface{})
	}
	if payload.Realm == "" {
		payload.Realm = password.DefaultRealm
	}
	return payload, nil
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

	// Create AuthInfo
	if err = h.AuthInfoStore.CreateAuth(&info); err != nil {
		if err == skydb.ErrUserDuplicated {
			err = ErrUserDuplicated
			return
		}

		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save auth info")
		return
	}

	// Create Profile
	var userProfile userprofile.UserProfile
	metadata := payload.Metadata
	if userProfile, err = h.UserProfileStore.CreateUserProfile(info.ID, metadata); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
		return
	}

	// Create Principal
	principals, err := h.createPrincipals(payload, info)
	if err != nil {
		return
	}
	loginPrincipal := principals[0]

	user := model.NewUser(info, userProfile)
	identities := []model.Identity{}
	var loginIdentity model.Identity
	for _, principal := range principals {
		identity := model.NewIdentity(h.IdentityProvider, principal)
		identities = append(identities, identity)
		if principal == loginPrincipal {
			loginIdentity = identity
		}
	}

	err = h.HookProvider.DispatchEvent(
		event.UserCreateEvent{
			User:       user,
			Identities: identities,
		},
		&user,
	)
	if err != nil {
		return
	}

	h.AuditTrail.Log(audit.Entry{
		AuthID: info.ID,
		Event:  audit.EventSignup,
	})

	// Create auth token
	tkn, err := h.TokenStore.NewToken(info.ID, loginPrincipal.PrincipalID())
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&tkn); err != nil {
		panic(err)
	}

	// Populate the activity time to user
	info.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	err = h.HookProvider.DispatchEvent(
		event.SessionCreateEvent{
			Reason:   event.SessionCreateReasonSignup,
			User:     user,
			Identity: loginIdentity,
		},
		&user,
	)
	if err != nil {
		return
	}

	resp = model.NewAuthResponse(user, loginIdentity, tkn.AccessToken)

	if h.WelcomeEmailEnabled {
		h.sendWelcomeEmail(user, payload.LoginIDs)
	}

	if h.AutoSendUserVerifyCode {
		h.sendUserVerifyRequest(user, payload.LoginIDs)
	}

	return
}

func (h SignupHandler) verifyPayload(payload SignupRequestPayload) (err error) {
	if err = h.PasswordAuthProvider.ValidateLoginIDs(payload.LoginIDs); err != nil {
		return
	}

	if valid := h.PasswordAuthProvider.IsRealmValid(payload.Realm); !valid {
		err = skyerr.NewInvalidArgument("realm is not allowed", []string{"realm"})
		return
	}

	// validate password
	err = h.PasswordChecker.ValidatePassword(authAudit.ValidatePasswordPayload{
		PlainPassword: payload.Password,
	})

	return
}

func (h SignupHandler) createPrincipals(payload SignupRequestPayload, authInfo authinfo.AuthInfo) (principals []principal.Principal, err error) {
	passwordPrincipals, createError := h.PasswordAuthProvider.CreatePrincipalsByLoginID(
		authInfo.ID,
		payload.Password,
		payload.LoginIDs,
		payload.Realm,
	)

	if createError != nil {
		if createError == skydb.ErrUserDuplicated {
			err = ErrUserDuplicated
		} else {
			err = createError
		}
	}
	if err == nil {
		for _, principal := range passwordPrincipals {
			principals = append(principals, principal)
		}
	}
	return
}

func (h SignupHandler) sendWelcomeEmail(user model.User, loginIDs []password.LoginID) {
	supportedLoginIDs := []password.LoginID{}
	for _, loginID := range loginIDs {
		if h.PasswordAuthProvider.CheckLoginIDKeyType(loginID.Key, metadata.Email) {
			supportedLoginIDs = append(supportedLoginIDs, loginID)
		}
	}

	var destinationLoginIDs []password.LoginID
	if h.WelcomeEmailDestination == config.WelcomeEmailDestinationAll {
		destinationLoginIDs = supportedLoginIDs
	} else if h.WelcomeEmailDestination == config.WelcomeEmailDestinationFirst {
		if len(supportedLoginIDs) > 0 {
			destinationLoginIDs = supportedLoginIDs[:1]
		}
	}

	for _, loginID := range destinationLoginIDs {
		email := loginID.Value
		h.TaskQueue.Enqueue(task.WelcomeEmailSendTaskName, task.WelcomeEmailSendTaskParam{
			Email: email,
			User:  user,
		}, nil)
	}
}

func (h SignupHandler) sendUserVerifyRequest(user model.User, loginIDs []password.LoginID) {
	for _, loginID := range loginIDs {
		for key := range h.UserVerifyLoginIDKeys {
			if key == loginID.Key {
				h.TaskQueue.Enqueue(task.VerifyCodeSendTaskName, task.VerifyCodeSendTaskParam{
					LoginID: loginID.Value,
					UserID:  user.ID,
				}, nil)
			}
		}
	}
}
