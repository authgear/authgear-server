package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
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
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/validation"
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
	inject.DefaultRequestInject(h, f.Dependency, request)
	h.AuditTrail = h.AuditTrail.WithRequest(request)
	return h.RequireAuthz(h, h)
}

type SignupRequestPayload struct {
	LoginIDs        []password.LoginID     `json:"login_ids"`
	Realm           string                 `json:"realm"`
	Password        string                 `json:"password"`
	Metadata        map[string]interface{} `json:"metadata"`
	OnUserDuplicate model.OnUserDuplicate  `json:"on_user_duplicate"`

	PasswordAuthProvider password.Provider          `json:"-"`
	AuthConfiguration    config.AuthConfiguration   `json:"-"`
	PasswordChecker      *authAudit.PasswordChecker `json:"-"`
}

// @JSONSchema
const SignupRequestSchema = `
{
	"$id": "#SignupRequest",
	"type": "object",
	"properties": {
		"login_ids": {
			"type": "array",
			"items": {
				"type": "object",
				"properties": {
					"key": { "type": "string", "minLength": 1 },
					"value": { "type": "string", "minLength": 1 }
				}
			},
			"minItems": 1
		},
		"realm": { "type": "string", "minLength": 1 },
		"password": { "type": "string", "minLength": 1 },
		"metadata": { "type": "object" },
		"on_user_duplicate": {
			"type": "string",
			"enum": ["abort", "create"]
		}
	}
}
`

func (p *SignupRequestPayload) SetDefaultValue() {
	if p.Realm == "" {
		p.Realm = password.DefaultRealm
	}
	if p.OnUserDuplicate == "" {
		p.OnUserDuplicate = model.OnUserDuplicateDefault
	}
	if p.Metadata == nil {
		// Avoid { metadata: null } in the response user object
		p.Metadata = make(map[string]interface{})
	}
}

func (p *SignupRequestPayload) Validate() []validation.ErrorCause {
	if !model.IsAllowedOnUserDuplicate(
		false,
		p.AuthConfiguration.OnUserDuplicateAllowCreate,
		p.OnUserDuplicate,
	) {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/on_user_duplicate",
			Message: "on_user_duplicate is not allowed",
		}}
	}

	if !p.PasswordAuthProvider.IsRealmValid(p.Realm) {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/realm",
			Message: "realm is not a valid realm",
		}}
	}

	loginIDs := map[string]struct{}{}

	for i, loginID := range p.LoginIDs {
		if _, found := loginIDs[loginID.Value]; found {
			return []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Pointer: fmt.Sprintf("/login_ids/%d/value", i),
				Message: "duplicated login ID",
			}}
		}
		loginIDs[loginID.Value] = struct{}{}
	}

	if err := p.PasswordAuthProvider.ValidateLoginIDs(p.LoginIDs); err != nil {
		return []validation.ErrorCause{{
			Kind:    validation.ErrorGeneral,
			Pointer: "/login_ids",
			Message: err.Error(),
		}}
	}

	return nil
}

/*
	@Operation POST /signup - Signup using password
		Signup user with login IDs and password.

		@Tag User

		@RequestBody
			Describe login IDs, password, and initial metadata.
			@JSONSchema {SignupRequest}

		@Response 200
			Signed up user and access token.
			@JSONSchema {AuthResponse}

		@Callback user_create {UserCreateEvent}
		@Callback session_create {SessionCreateEvent}
		@Callback user_sync {UserSyncEvent}
*/
type SignupHandler struct {
	RequireAuthz            handler.RequireAuthz                               `dependency:"RequireAuthz"`
	Validator               *validation.Validator                              `dependency:"Validator"`
	AuthnSessionProvider    authnsession.Provider                              `dependency:"AuthnSessionProvider"`
	PasswordChecker         *authAudit.PasswordChecker                         `dependency:"PasswordChecker"`
	UserProfileStore        userprofile.Store                                  `dependency:"UserProfileStore"`
	AuthInfoStore           authinfo.Store                                     `dependency:"AuthInfoStore"`
	PasswordAuthProvider    password.Provider                                  `dependency:"PasswordAuthProvider"`
	IdentityProvider        principal.IdentityProvider                         `dependency:"IdentityProvider"`
	AuditTrail              audit.Trail                                        `dependency:"AuditTrail"`
	WelcomeEmailEnabled     bool                                               `dependency:"WelcomeEmailEnabled"`
	WelcomeEmailDestination config.WelcomeEmailDestination                     `dependency:"WelcomeEmailDestination"`
	AutoSendUserVerifyCode  bool                                               `dependency:"AutoSendUserVerifyCodeOnSignup"`
	UserVerifyLoginIDKeys   map[string]config.UserVerificationKeyConfiguration `dependency:"UserVerifyLoginIDKeys"`
	AuthConfiguration       config.AuthConfiguration                           `dependency:"AuthConfiguration"`
	TxContext               db.TxContext                                       `dependency:"TxContext"`
	Logger                  *logrus.Entry                                      `dependency:"HandlerLogger"`
	TaskQueue               async.Queue                                        `dependency:"AsyncTaskQueue"`
	HookProvider            hook.Provider                                      `dependency:"HookProvider"`
	URLPrefix               *url.URL                                           `dependency:"URLPrefix"`
}

func (h SignupHandler) ProvideAuthzPolicy() authz.Policy {
	return authz.PolicyFunc(policy.DenyNoAccessKey)
}

func (h SignupHandler) DecodeRequest(request *http.Request, resp http.ResponseWriter) (payload SignupRequestPayload, err error) {
	payload.PasswordAuthProvider = h.PasswordAuthProvider
	payload.AuthConfiguration = h.AuthConfiguration
	err = handler.BindJSONBody(request, resp, h.Validator, "#SignupRequest", &payload)
	return
}

func (h SignupHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error

	payload, err := h.DecodeRequest(req, resp)
	if err != nil {
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

func (h SignupHandler) Handle(payload SignupRequestPayload) (resp interface{}, err error) {
	// validate password
	if err = h.PasswordChecker.ValidatePassword(authAudit.ValidatePasswordPayload{
		PlainPassword: payload.Password,
	}); err != nil {
		return
	}

	existingPrincipals, err := h.findExistingPrincipals(payload)
	if err != nil {
		return
	}

	if len(existingPrincipals) > 0 && payload.OnUserDuplicate == model.OnUserDuplicateAbort {
		err = password.ErrLoginIDAlreadyUsed
		return
	}

	now := timeNow()
	info := authinfo.NewAuthInfo()
	info.LastLoginAt = &now

	// Create AuthInfo
	if err = h.AuthInfoStore.CreateAuth(&info); err != nil {
		return
	}

	// Create Profile
	userProfile, err := h.UserProfileStore.CreateUserProfile(info.ID, payload.Metadata)
	if err != nil {
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
	for _, principal := range principals {
		identity := model.NewIdentity(h.IdentityProvider, principal)
		identities = append(identities, identity)
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
		UserID: info.ID,
		Event:  audit.EventSignup,
	})

	sess, err := h.AuthnSessionProvider.NewFromScratch(info.ID, loginPrincipal, coreAuth.SessionCreateReasonSignup)
	if err != nil {
		return
	}
	resp, err = h.AuthnSessionProvider.GenerateResponseAndUpdateLastLoginAt(*sess)
	if err != nil {
		return
	}

	if h.WelcomeEmailEnabled {
		h.sendWelcomeEmail(user, payload.LoginIDs)
	}

	if h.AutoSendUserVerifyCode {
		h.sendUserVerifyRequest(user, payload.LoginIDs)
	}

	return
}

func (h SignupHandler) findExistingPrincipals(payload SignupRequestPayload) ([]principal.Principal, error) {
	var principals []principal.Principal

	// Find out all login IDs that are of type email.
	var emails []string
	for _, loginID := range payload.LoginIDs {
		if h.PasswordAuthProvider.CheckLoginIDKeyType(loginID.Key, metadata.Email) {
			emails = append(emails, loginID.Value)
		}
	}

	// For each email, find out all principals.
	for _, email := range emails {
		ps, err := h.IdentityProvider.ListPrincipalsByClaim("email", email)
		if err != nil {
			return nil, err
		}
		principals = append(principals, ps...)
	}

	// Skip password principals which are not in the same realm.
	var filteredPrincipals []principal.Principal
	for _, p := range principals {
		if passwordPrincipal, ok := p.(*password.Principal); ok && passwordPrincipal.Realm != payload.Realm {
			continue
		}
		filteredPrincipals = append(filteredPrincipals, p)
	}

	return filteredPrincipals, nil
}

func (h SignupHandler) createPrincipals(payload SignupRequestPayload, authInfo authinfo.AuthInfo) (principals []principal.Principal, err error) {
	passwordPrincipals, err := h.PasswordAuthProvider.CreatePrincipalsByLoginID(
		authInfo.ID,
		payload.Password,
		payload.LoginIDs,
		payload.Realm,
	)
	if err != nil {
		return
	}

	for _, principal := range passwordPrincipals {
		principals = append(principals, principal)
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
			URLPrefix: h.URLPrefix,
			Email:     email,
			User:      user,
		}, nil)
	}
}

func (h SignupHandler) sendUserVerifyRequest(user model.User, loginIDs []password.LoginID) {
	for _, loginID := range loginIDs {
		for key := range h.UserVerifyLoginIDKeys {
			if key == loginID.Key {
				h.TaskQueue.Enqueue(task.VerifyCodeSendTaskName, task.VerifyCodeSendTaskParam{
					URLPrefix: h.URLPrefix,
					LoginID:   loginID.Value,
					UserID:    user.ID,
				}, nil)
			}
		}
	}
}
