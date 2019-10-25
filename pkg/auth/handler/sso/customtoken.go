package sso

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/customtoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
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
	return h.RequireAuthz(h, h)
}

type CustomTokenLoginPayload struct {
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

func (payload CustomTokenLoginPayload) Validate() error {
	// TODO(error): JSON schema
	if err := payload.Claims.Validate(payload.ExpectedIssuer, payload.ExpectedAudience); err != nil {
		return sso.NewSSOFailed(sso.SSOUnauthorized, "invalid token")
	}

	if !model.IsValidOnUserDuplicateForSSO(payload.OnUserDuplicate) {
		return skyerr.NewInvalid("invalid OnUserDuplicate")
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
	RequireAuthz             handler.RequireAuthz            `dependency:"RequireAuthz"`
	AuthnSessionProvider     authnsession.Provider           `dependency:"AuthnSessionProvider"`
	UserProfileStore         userprofile.Store               `dependency:"UserProfileStore"`
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

// ProvideAuthzPolicy provides authorization policy of handler
func (h CustomTokenLoginHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

// DecodeRequest decode request payload
func (h CustomTokenLoginHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload CustomTokenLoginPayload, err error) {
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

	if err = handler.DecodeJSONBody(request, resp, &payload); err != nil {
		return
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
		err = sso.NewSSOFailed(sso.SSOUnauthorized, "invalid token")
		return
	}
	return
}

func (h CustomTokenLoginHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
		h.AuthnSessionProvider.WriteResponse(resp, nil, err)
		return
	}

	if err = payload.Validate(); err != nil {
		h.AuthnSessionProvider.WriteResponse(resp, nil, err)
		return
	}

	result, err := handler.Transactional(h.TxContext, func() (result interface{}, err error) {
		result, err = h.Handle(payload)
		if err == nil {
			err = h.HookProvider.WillCommitTx()
		}
		return
	})
	if err == nil {
		h.HookProvider.DidCommitTx()
	}
	h.AuthnSessionProvider.WriteResponse(resp, result, err)
}

// Handle function handle custom token login
func (h CustomTokenLoginHandler) Handle(payload CustomTokenLoginPayload) (resp interface{}, err error) {
	if !h.CustomTokenConfiguration.Enabled {
		err = skyerr.NewNotFound("custom token is disabled")
		return
	}

	// TODO(error): JSON schema
	if !h.PasswordAuthProvider.IsRealmValid(payload.MergeRealm) {
		err = skyerr.NewInvalid("invalid MergeRealm")
		return
	}

	if !model.IsAllowedOnUserDuplicate(
		h.CustomTokenConfiguration.OnUserDuplicateAllowMerge,
		h.CustomTokenConfiguration.OnUserDuplicateAllowCreate,
		payload.OnUserDuplicate,
	) {
		err = skyerr.NewInvalid("disallowed OnUserDuplicate")
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
		return
	}

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

	var sessionCreateReason coreAuth.SessionCreateReason
	if createNewUser {
		sessionCreateReason = coreAuth.SessionCreateReasonSignup
	} else {
		sessionCreateReason = coreAuth.SessionCreateReasonLogin
	}

	sess, err := h.AuthnSessionProvider.NewFromScratch(principal.UserID, principal, sessionCreateReason)
	if err != nil {
		return
	}
	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*sess)
	if err != nil {
		return
	}

	// TODO: audit trail
	if createNewUser && h.WelcomeEmailEnabled {
		h.sendWelcomeEmail(user, payload.Claims.Email())
	}

	return
}

func (h CustomTokenLoginHandler) handleLogin(
	payload CustomTokenLoginPayload,
	info *authinfo.AuthInfo,
) (createNewUser bool, customTokenPrincipal *customtoken.Principal, err error) {
	customTokenPrincipal, err = h.findExistingCustomTokenPrincipal(payload.Claims.Subject())
	if err != nil && !errors.Is(err, principal.ErrNotFound) {
		return
	}

	now := timeNow()

	createFunc := func() {
		createNewUser = true

		*info = authinfo.NewAuthInfo()
		info.LastLoginAt = &now

		// Create AuthInfo
		if e := h.AuthInfoStore.CreateAuth(info); e != nil {
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

		err = h.AuthInfoStore.GetAuth(customTokenPrincipal.UserID, info)
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
		err = password.ErrLoginIDAlreadyUsed
	case model.OnUserDuplicateCreate:
		createFunc()
	case model.OnUserDuplicateMerge:
		// Case: The same email is shared by multiple users
		if len(userIDs) > 1 {
			err = password.ErrLoginIDAlreadyUsed
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
		err = h.AuthInfoStore.GetAuth(userID, info)
	}

	return
}

func (h CustomTokenLoginHandler) findExistingCustomTokenPrincipal(subject string) (*customtoken.Principal, error) {
	principal, err := h.CustomTokenAuthProvider.GetPrincipalByTokenPrincipalID(subject)
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
