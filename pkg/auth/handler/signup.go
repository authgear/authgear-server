package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/utils"

	"github.com/skygeario/skygear-server/pkg/auth/task"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/response"
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
	h.HookStore = h.HookStore.WithRequest(request)
	return auth.HookHandlerToHandler(h, h.TxContext)
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
	if p.isAnonymous() {
		//no validation logic for anonymous sign up
	} else {
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

func (p SignupRequestPayload) isAnonymous() bool {
	return len(p.LoginIDs) == 0 && p.Password == ""
}

// SignupHandler handles signup request
type SignupHandler struct {
	PasswordChecker         *authAudit.PasswordChecker                         `dependency:"PasswordChecker"`
	UserProfileStore        userprofile.Store                                  `dependency:"UserProfileStore"`
	TokenStore              authtoken.Store                                    `dependency:"TokenStore"`
	AuthInfoStore           authinfo.Store                                     `dependency:"AuthInfoStore"`
	PasswordAuthProvider    password.Provider                                  `dependency:"PasswordAuthProvider"`
	AnonymousAuthProvider   anonymous.Provider                                 `dependency:"AnonymousAuthProvider"`
	AuditTrail              audit.Trail                                        `dependency:"AuditTrail"`
	WelcomeEmailEnabled     bool                                               `dependency:"WelcomeEmailEnabled"`
	WelcomeEmailDestination config.WelcomeEmailDestination                     `dependency:"WelcomeEmailDestination"`
	AutoSendUserVerifyCode  bool                                               `dependency:"AutoSendUserVerifyCodeOnSignup"`
	UserVerifyKeys          map[string]config.UserVerificationKeyConfiguration `dependency:"UserVerifyKeys"`
	VerifyCodeStore         userverify.Store                                   `dependency:"VerifyCodeStore"`
	TxContext               db.TxContext                                       `dependency:"TxContext"`
	Logger                  *logrus.Entry                                      `dependency:"HandlerLogger"`
	TaskQueue               async.Queue                                        `dependency:"AsyncTaskQueue"`
	HookStore               hook.Store                                         `dependency:"HookStore"`
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

func (h SignupHandler) ExecBeforeHooks(req interface{}, inputUser *response.User) error {
	payload := req.(SignupRequestPayload)
	inputUser.Metadata = payload.Metadata
	err := h.HookStore.ExecBeforeHooksByEvent(hook.BeforeSignup, req, inputUser, "")
	return err
}

func (h SignupHandler) HandleRequest(req interface{}, inputUser *response.User) (resp interface{}, err error) {
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
	if inputUser != nil {
		metadata = inputUser.Metadata
	}
	if userProfile, err = h.UserProfileStore.CreateUserProfile(info.ID, metadata); err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
		return
	}

	// Create Principal
	if err = h.createPrincipal(payload, info); err != nil {
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

	userFactory := response.UserFactory{
		PasswordAuthProvider: h.PasswordAuthProvider,
	}
	user := userFactory.NewUser(info, userProfile)

	// Populate the activity time to user
	info.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	h.AuditTrail.Log(audit.Entry{
		AuthID: info.ID,
		Event:  audit.EventSignup,
	})

	*inputUser = user

	resp = response.NewAuthResponseByUser(user, tkn.AccessToken)

	return
}

func (h SignupHandler) ExecAfterHooks(req interface{}, resp interface{}, user response.User) error {
	reqPayload := req.(SignupRequestPayload)
	respPayload := resp.(response.AuthResponse)
	err := h.HookStore.ExecAfterHooksByEvent(hook.AfterSignup, req, user, respPayload.AccessToken)
	if err != nil {
		return err
	}

	if h.WelcomeEmailEnabled {
		h.sendWelcomeEmail(user, reqPayload.LoginIDs)
	}

	if h.AutoSendUserVerifyCode {
		h.sendUserVerifyRequest(user, reqPayload.LoginIDs)
	}

	return nil
}

func (h SignupHandler) verifyPayload(payload SignupRequestPayload) (err error) {
	if payload.isAnonymous() {
		return
	}

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

func (h SignupHandler) createPrincipal(payload SignupRequestPayload, authInfo authinfo.AuthInfo) (err error) {
	if !payload.isAnonymous() {
		err = h.PasswordAuthProvider.CreatePrincipalsByLoginID(authInfo.ID, payload.Password, payload.LoginIDs, payload.Realm)
		if err == skydb.ErrUserDuplicated {
			err = ErrUserDuplicated
		}
	} else {
		principal := anonymous.NewPrincipal()
		principal.UserID = authInfo.ID

		err = h.AnonymousAuthProvider.CreatePrincipal(principal)
	}

	return
}

func (h SignupHandler) sendWelcomeEmail(user response.User, loginIDs []password.LoginID) {
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

func (h SignupHandler) sendUserVerifyRequest(user response.User, loginIDs []password.LoginID) {
	for _, loginID := range loginIDs {
		for key := range h.UserVerifyKeys {
			if key == loginID.Key {
				h.TaskQueue.Enqueue(task.VerifyCodeSendTaskName, task.VerifyCodeSendTaskParam{
					LoginIDKey: loginID.Key,
					LoginID:    loginID.Value,
					User:       user,
				}, nil)
			}
		}
	}
}
