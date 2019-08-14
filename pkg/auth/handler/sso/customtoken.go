package sso

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/event"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/customtoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	signUpHandler "github.com/skygeario/skygear-server/pkg/auth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// AttachCustomTokenLoginHandler attaches CustomTokenLoginHandler to server
func AttachCustomTokenLoginHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/sso/custom_token/login", &CustomTokenLoginHandlerFactory{
		authDependency,
	}).Methods("OPTIONS", "POST")
	return server
}

// CustomTokenLoginHandlerFactory creates CustomTokenLoginHandler
type CustomTokenLoginHandlerFactory struct {
	Dependency auth.DependencyMap
}

// NewHandler creates new CustomTokenLoginHandler
func (f CustomTokenLoginHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &CustomTokenLoginHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return handler.APIHandlerToHandler(hook.WrapHandler(h.HookProvider, h), h.TxContext)
}

// ProvideAuthzPolicy provides authorization policy of handler
func (f CustomTokenLoginHandlerFactory) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

type customTokenLoginPayload struct {
	TokenString      string                           `json:"token"`
	MergeRealm       string                           `json:"merge_realm"`
	OnUserDuplicate  model.OnUserDuplicate            `json:"on_user_duplicate"`
	Claims           customtoken.SSOCustomTokenClaims `json:"-"`
	ExpectedIssuer   string                           `json:"-"`
	ExpectedAudience string                           `json:"-"`
}

// nolint: gosec
// @JSONSchema
const CustomTokenLoginRequestSchema = `
{
	"$id": "#CustomTokenLoginRequest",
	"type": "object",
	"properties": {
		"token": { "type": "string" },
		"merge_realm": { "type": "string" },
		"on_user_duplicate": { "type": "string" }
	}
}
`

func (payload customTokenLoginPayload) Validate() error {
	if err := payload.Claims.Validate(payload.ExpectedIssuer, payload.ExpectedAudience); err != nil {
		return skyerr.NewError(
			skyerr.InvalidCredentials,
			err.Error(),
		)
	}

	if !model.IsValidOnUserDuplicateForSSO(payload.OnUserDuplicate) {
		return skyerr.NewInvalidArgument("Invalid OnUserDuplicate", []string{"on_user_duplicate"})
	}

	return nil
}

// nolint: gosec
/*
	@Operation POST /sso/custom_token/login - Authenticate with custom token
		An external server is responsible for generating the custom token which
		contains a Principal ID and a signature. It is required that the token
		has issued-at and expired-at claims.

		The custom token is signed by a shared secret and encoded in JWT format.

		The claims of the custom token is as follows:

			{
			"sub": "id1234567800",
			"iat": 1513316033,
			"exp": 1828676033,
			"email": "johndoe@oursky.com"
			}

		When signing the above claims with the custom token secret `ssosecret` using
		HS256 as algorithm, the following JWT token is produced:

			eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJpZDEyMzQ1Njc4MDAiLCJpYXQiOjE1MTMzMTYwMzMsImV4cCI6MTgyODY3NjAzMywic2t5cHJvZmlsZSI6eyJuYW1lIjoiSm9obiBEb2UifX0.JRAwXPF4CDWCpMCvemCBPrUAQAXPV9qVWeAYo1vBAqQ

		This token can be used to log in to Skygear Server. If there is no user
		associated with the Token Principal ID (the subject/sub claim), a new user is
		created.

		@Tag SSO

		@RequestBody
			@JSONSchema {CustomTokenLoginRequest}

		@Response 200
			Logged in user and access token.
			@JSONSchema {AuthResponse}

		@Callback user_create {UserSyncEvent}
		@Callback identity_create {UserSyncEvent}
		@Callback session_create {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type CustomTokenLoginHandler struct {
	TxContext                db.TxContext                    `dependency:"TxContext"`
	UserProfileStore         userprofile.Store               `dependency:"UserProfileStore"`
	TokenStore               authtoken.Store                 `dependency:"TokenStore"`
	AuthInfoStore            authinfo.Store                  `dependency:"AuthInfoStore"`
	CustomTokenAuthProvider  customtoken.Provider            `dependency:"CustomTokenAuthProvider"`
	IdentityProvider         principal.IdentityProvider      `dependency:"IdentityProvider"`
	HookProvider             hook.Provider                   `dependency:"HookProvider"`
	PasswordAuthProvider     password.Provider               `dependency:"PasswordAuthProvider"`
	CustomTokenConfiguration config.CustomTokenConfiguration `dependency:"CustomTokenConfiguration"`
	WelcomeEmailEnabled      bool                            `dependency:"WelcomeEmailEnabled"`
	AuditTrail               audit.Trail                     `dependency:"AuditTrail"`
	TaskQueue                async.Queue                     `dependency:"AsyncTaskQueue"`
}

func (h CustomTokenLoginHandler) WithTx() bool {
	return true
}

// DecodeRequest decode request payload
func (h CustomTokenLoginHandler) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	payload := customTokenLoginPayload{}
	var err error

	defer func() {
		if err != nil {
			h.AuditTrail.Log(audit.Entry{
				Event: audit.EventLoginFailure,
				Data: map[string]interface{}{
					"type": "custom_token",
				},
			})
		}
	}()

	if err = json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return nil, err
	}

	payload.ExpectedIssuer = h.CustomTokenConfiguration.Issuer
	payload.ExpectedAudience = h.CustomTokenConfiguration.Audience

	if payload.MergeRealm == "" {
		payload.MergeRealm = password.DefaultRealm
	}

	if payload.OnUserDuplicate == "" {
		payload.OnUserDuplicate = model.OnUserDuplicateDefault
	}

	payload.Claims, err = h.CustomTokenAuthProvider.Decode(payload.TokenString)
	if err != nil {
		return nil, skyerr.NewError(skyerr.BadRequest, err.Error())
	}
	return payload, err
}

// Handle function handle custom token login
func (h CustomTokenLoginHandler) Handle(req interface{}) (resp interface{}, err error) {
	if !h.CustomTokenConfiguration.Enabled {
		err = skyerr.NewError(skyerr.UndefinedOperation, "Custom Token is disabled")
		return
	}

	payload := req.(customTokenLoginPayload)

	if !h.PasswordAuthProvider.IsRealmValid(payload.MergeRealm) {
		err = skyerr.NewInvalidArgument("Invalid MergeRealm", []string{payload.MergeRealm})
		return
	}

	if !model.IsAllowedOnUserDuplicate(
		h.CustomTokenConfiguration.OnUserDuplicateAllowMerge,
		h.CustomTokenConfiguration.OnUserDuplicateAllowCreate,
		payload.OnUserDuplicate,
	) {
		err = skyerr.NewInvalidArgument("Disallowed OnUserDuplicate", []string{string(payload.OnUserDuplicate)})
		return
	}

	var info authinfo.AuthInfo
	var createNewUser bool

	defer func() {
		var event audit.Event
		if err != nil {
			event = audit.EventLoginFailure
		} else {
			if createNewUser {
				event = audit.EventSignup
			} else {
				event = audit.EventLoginSuccess
			}
		}

		h.AuditTrail.Log(audit.Entry{
			UserID: info.ID,
			Event:  event,
			Data: map[string]interface{}{
				"type": "custom_token",
			},
		})
	}()

	createNewUser, principal, err := h.handleLogin(payload, &info)
	if err != nil {
		return
	}

	// Create empty user profile
	var userProfile userprofile.UserProfile
	emptyProfile := map[string]interface{}{}
	if createNewUser {
		userProfile, err = h.UserProfileStore.CreateUserProfile(info.ID, emptyProfile)
	} else {
		userProfile, err = h.UserProfileStore.GetUserProfile(info.ID)
	}
	if err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
		return
	}

	// TODO: check disable

	user := model.NewUser(info, userProfile)
	identity := model.NewIdentity(h.IdentityProvider, principal)

	if createNewUser {
		err = h.HookProvider.DispatchEvent(
			event.UserCreateEvent{
				User:       user,
				Identities: []model.Identity{identity},
			},
			&user,
		)
		if err != nil {
			return
		}
	}

	// Create auth token
	tkn, err := h.TokenStore.NewToken(info.ID, principal.ID)
	if err != nil {
		panic(err)
	}

	if err = h.TokenStore.Put(&tkn); err != nil {
		panic(err)
	}

	var sessionCreateReason event.SessionCreateReason
	if createNewUser {
		sessionCreateReason = event.SessionCreateReasonSignup
	} else {
		sessionCreateReason = event.SessionCreateReasonLogin
	}
	err = h.HookProvider.DispatchEvent(
		event.SessionCreateEvent{
			Reason:   sessionCreateReason,
			User:     user,
			Identity: identity,
		},
		&user,
	)
	if err != nil {
		return
	}

	// Reload auth info, in case before hook handler mutated it
	if err = h.AuthInfoStore.GetAuth(principal.UserID, &info); err != nil {
		return
	}

	// Update the activity time of user (return old activity time for usefulness)
	now := timeNow()
	info.LastLoginAt = &now
	info.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	resp = model.NewAuthResponse(user, identity, tkn.AccessToken)

	// TODO: audit trail
	if createNewUser && h.WelcomeEmailEnabled {
		h.sendWelcomeEmail(user, payload.Claims.Email())
	}

	return
}

func (h CustomTokenLoginHandler) handleLogin(
	payload customTokenLoginPayload,
	info *authinfo.AuthInfo,
) (createNewUser bool, customTokenPrincipal *customtoken.Principal, err error) {
	customTokenPrincipal, err = h.findExistingCustomTokenPrincipal(payload.Claims.Subject())
	if err != nil {
		return
	}

	now := timeNow()

	populateInfo := func(userID string) {
		if e := h.AuthInfoStore.GetAuth(userID, info); e != nil {
			if e == skydb.ErrUserNotFound {
				err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
				return
			}
			return
		}
	}

	createFunc := func() {
		createNewUser = true

		*info = authinfo.NewAuthInfo()
		info.LastLoginAt = &now

		// Create AuthInfo
		if e := h.AuthInfoStore.CreateAuth(info); e != nil {
			if e == skydb.ErrUserDuplicated {
				err = signUpHandler.ErrUserDuplicated
				return
			}

			// TODO:
			// return proper error
			err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save auth info")
			return
		}

		customTokenPrincipal, err = h.createCustomTokenPrincipal(info.ID, payload.Claims)
		if err != nil {
			return
		}
	}

	// Case: Custom Token principal was found
	// => Simple update case
	// We do not need to consider other principals
	if customTokenPrincipal != nil {
		customTokenPrincipal.SetRawProfile(payload.Claims)
		err = h.CustomTokenAuthProvider.UpdatePrincipal(customTokenPrincipal)
		if err != nil {
			return
		}

		populateInfo(customTokenPrincipal.UserID)
		return
	}

	// Case: Custom Token principal was not found
	// We need to consider other principals
	principals, err := h.findExistingPrincipals(payload.Claims.Email(), payload.MergeRealm)
	if err != nil {
		return
	}
	userIDs := h.principalsToUserIDs(principals)

	// Case: Custom Token principal was not found and no other principals were not found
	// => Simple create case
	if len(userIDs) <= 0 {
		createFunc()
		return
	}

	// Case: Custom Token principal was not found and Password principal was found
	// => Complex case
	switch payload.OnUserDuplicate {
	case model.OnUserDuplicateAbort:
		err = skyerr.NewError(skyerr.Duplicated, "Aborted due to duplicate user")
	case model.OnUserDuplicateCreate:
		createFunc()
	case model.OnUserDuplicateMerge:
		// Case: The same email is shared by multiple users
		if len(userIDs) > 1 {
			err = skyerr.NewError(skyerr.Duplicated, "Email shared by multiple users")
			return
		}
		// Associate the provider to the existing user
		userID := userIDs[0]
		customTokenPrincipal, err = h.createCustomTokenPrincipal(
			userID,
			payload.Claims,
		)
		if err != nil {
			return
		}
		populateInfo(userID)
	}

	return
}

func (h CustomTokenLoginHandler) findExistingCustomTokenPrincipal(subject string) (*customtoken.Principal, error) {
	principal, err := h.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID(subject)
	if err == skydb.ErrUserNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return principal, nil
}

func (h CustomTokenLoginHandler) sendWelcomeEmail(user model.User, email string) {
	if email != "" {
		h.TaskQueue.Enqueue(task.WelcomeEmailSendTaskName, task.WelcomeEmailSendTaskParam{
			Email: email,
			User:  user,
		}, nil)
	}
}

func (h CustomTokenLoginHandler) findExistingPrincipals(email string, mergeRealm string) ([]principal.Principal, error) {
	if email == "" {
		return nil, nil
	}
	principals, err := h.IdentityProvider.ListPrincipalsByClaim("email", email)
	if err != nil {
		return nil, err
	}
	var filteredPrincipals []principal.Principal
	for _, p := range principals {
		if passwordPrincipal, ok := p.(*password.Principal); ok {
			if passwordPrincipal.Realm == mergeRealm {
				filteredPrincipals = append(filteredPrincipals, p)
			}
		} else {
			filteredPrincipals = append(filteredPrincipals, p)
		}
	}
	return filteredPrincipals, nil
}

func (h CustomTokenLoginHandler) principalsToUserIDs(principals []principal.Principal) []string {
	seen := map[string]struct{}{}
	var userIDs []string
	for _, p := range principals {
		userID := p.PrincipalUserID()
		_, ok := seen[userID]
		if !ok {
			seen[userID] = struct{}{}
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

func (h CustomTokenLoginHandler) createCustomTokenPrincipal(userID string, claims customtoken.SSOCustomTokenClaims) (*customtoken.Principal, error) {
	customTokenPrincipal := customtoken.NewPrincipal()
	customTokenPrincipal.TokenPrincipalID = claims.Subject()
	customTokenPrincipal.UserID = userID
	customTokenPrincipal.SetRawProfile(claims)
	err := h.CustomTokenAuthProvider.CreatePrincipal(&customTokenPrincipal)
	return &customTokenPrincipal, err
}
