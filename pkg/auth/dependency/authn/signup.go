package authn

import (
	"fmt"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	task "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func principalsToUserIDs(principals []principal.Principal) []string {
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

// SignupProcess handle user creation: create principals (and user if needed) according to provided information
type SignupProcess struct {
	PasswordChecker               *audit.PasswordChecker
	LoginIDChecker                loginid.LoginIDChecker
	IdentityProvider              principal.IdentityProvider
	PasswordProvider              password.Provider
	OAuthProvider                 oauth.Provider
	TimeProvider                  coreTime.Provider
	AuthInfoStore                 authinfo.Store
	UserProfileStore              userprofile.Store
	HookProvider                  hook.Provider
	WelcomeEmailConfiguration     *config.WelcomeEmailConfiguration
	UserVerificationConfiguration *config.UserVerificationConfiguration
	AuthConfiguration             *config.AuthConfiguration
	URLPrefixProvider             urlprefix.Provider
	TaskQueue                     async.Queue
}

func (p *SignupProcess) ValidateSignupLoginID(loginID loginid.LoginID) (err error) {
	err = p.validateCreateUserWithLoginIDs([]loginid.LoginID{loginID}, model.OnUserDuplicateAbort)
	return
}

func (p *SignupProcess) SignupWithLoginIDs(
	loginIDs []loginid.LoginID,
	plainPassword string,
	metadata map[string]interface{},
	onUserDuplicate model.OnUserDuplicate,
) (firstPrincipal principal.Principal, err error) {
	err = p.validateCreateUserWithLoginIDs(
		loginIDs,
		onUserDuplicate,
	)
	if err != nil {
		return
	}

	err = p.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: plainPassword,
	})
	if err != nil {
		return
	}

	existingPrincipals, err := p.findExistingPrincipalsWithProposedLoginIDs(loginIDs)
	if err != nil {
		return
	}

	if len(existingPrincipals) > 0 && onUserDuplicate == model.OnUserDuplicateAbort {
		err = password.ErrLoginIDAlreadyUsed
		return
	}

	now := p.TimeProvider.NowUTC()
	authInfoS := authinfo.NewAuthInfo()
	authInfo := &authInfoS
	authInfo.LastLoginAt = &now

	err = p.AuthInfoStore.CreateAuth(authInfo)
	if err != nil {
		return
	}

	userProfileS, err := p.UserProfileStore.CreateUserProfile(authInfo.ID, metadata)
	if err != nil {
		return
	}
	userProfile := &userProfileS

	principals, err := p.createPrincipalsWithProposedLoginIDs(authInfo.ID, plainPassword, loginIDs)
	if err != nil {
		return
	}
	firstPrincipal = principals[0]

	user := model.NewUser(*authInfo, *userProfile)
	identities := []model.Identity{}
	for _, principal := range principals {
		identity := model.NewIdentity(principal)
		identities = append(identities, identity)
	}

	err = p.HookProvider.DispatchEvent(
		event.UserCreateEvent{
			User:       user,
			Identities: identities,
		},
		&user,
	)
	if err != nil {
		return
	}

	if p.WelcomeEmailConfiguration.Enabled {
		p.enqueueSendWelcomeEmailTasks(user, loginIDs)
	}

	if p.UserVerificationConfiguration.AutoSendOnSignup {
		p.enqueueSendVerificationCodeTasks(user, loginIDs)
	}

	return
}

func (p *SignupProcess) validateCreateUserWithLoginIDs(loginIDs []loginid.LoginID, onUserDuplicate model.OnUserDuplicate) (err error) {
	var causes []validation.ErrorCause

	if !model.IsAllowedOnUserDuplicate(
		false,
		p.AuthConfiguration.OnUserDuplicateAllowCreate,
		onUserDuplicate,
	) {
		causes = append(causes, validation.ErrorCause{
			Kind:    validation.ErrorGeneral,
			Pointer: "/on_user_duplicate",
			Message: "on_user_duplicate is not allowed",
		})
	}

	seen := map[string]struct{}{}

	for i, loginID := range loginIDs {
		if _, found := seen[loginID.Value]; found {
			causes = append(causes, validation.ErrorCause{
				Kind:    validation.ErrorGeneral,
				Pointer: fmt.Sprintf("/login_ids/%d/value", i),
				Message: "duplicated login ID",
			})
		}
		seen[loginID.Value] = struct{}{}
	}

	if err := p.LoginIDChecker.Validate(loginIDs); err != nil {
		if cs := validation.ErrorCauses(err); len(cs) > 0 {
			for i := range cs {
				cs[i].Pointer = fmt.Sprintf("/login_ids%s", cs[i].Pointer)
			}
			causes = append(causes, cs...)
		}
	}

	if len(causes) > 0 {
		err = validation.NewValidationFailed("invalid request body", causes)
		return
	}

	return nil
}

func (p *SignupProcess) findExistingPrincipalsWithProposedLoginIDs(loginIDs []loginid.LoginID) (principals []principal.Principal, err error) {
	// Find out all login IDs that are of type email.
	var emails []string
	for _, loginID := range loginIDs {
		if p.LoginIDChecker.CheckType(loginID.Key, metadata.Email) {
			emails = append(emails, loginID.Value)
		}
	}

	// For each email, find out all principals.
	var ps []principal.Principal
	for _, email := range emails {
		ps, err = p.IdentityProvider.ListPrincipalsByClaim("email", email)
		if err != nil {
			return
		}
		principals = append(principals, ps...)
	}

	realm := password.DefaultRealm
	var filteredPrincipals []principal.Principal
	for _, p := range principals {
		if passwordPrincipal, ok := p.(*password.Principal); ok && passwordPrincipal.Realm != realm {
			continue
		}
		filteredPrincipals = append(filteredPrincipals, p)
	}

	principals = filteredPrincipals
	return
}

func (p *SignupProcess) createPrincipalsWithProposedLoginIDs(userID string, plainPassword string, loginIDs []loginid.LoginID) (principals []principal.Principal, err error) {
	realm := password.DefaultRealm
	passwordPrincipals, err := p.PasswordProvider.CreatePrincipalsByLoginID(
		userID,
		plainPassword,
		loginIDs,
		realm,
	)
	if err != nil {
		return
	}

	for _, principal := range passwordPrincipals {
		principals = append(principals, principal)
	}
	return
}

func (p *SignupProcess) enqueueSendWelcomeEmailTasks(user model.User, loginIDs []loginid.LoginID) {
	supportedLoginIDs := []loginid.LoginID{}
	for _, loginID := range loginIDs {
		if p.LoginIDChecker.CheckType(loginID.Key, metadata.Email) {
			supportedLoginIDs = append(supportedLoginIDs, loginID)
		}
	}

	var destinationLoginIDs []loginid.LoginID
	if p.WelcomeEmailConfiguration.Destination == config.WelcomeEmailDestinationAll {
		destinationLoginIDs = supportedLoginIDs
	} else if p.WelcomeEmailConfiguration.Destination == config.WelcomeEmailDestinationFirst {
		if len(supportedLoginIDs) > 0 {
			destinationLoginIDs = supportedLoginIDs[:1]
		}
	}

	for _, loginID := range destinationLoginIDs {
		email := loginID.Value
		p.TaskQueue.Enqueue(async.TaskSpec{
			Name: task.WelcomeEmailSendTaskName,
			Param: task.WelcomeEmailSendTaskParam{
				URLPrefix: p.URLPrefixProvider.Value(),
				Email:     email,
				User:      user,
			},
		})
	}

	return
}

func (p *SignupProcess) enqueueSendVerificationCodeTasks(user model.User, loginIDs []loginid.LoginID) {
	for _, loginID := range loginIDs {
		for _, keyConfig := range p.UserVerificationConfiguration.LoginIDKeys {
			if keyConfig.Key == loginID.Key {
				p.TaskQueue.Enqueue(async.TaskSpec{
					Name: task.VerifyCodeSendTaskName,
					Param: task.VerifyCodeSendTaskParam{
						URLPrefix: p.URLPrefixProvider.Value(),
						LoginID:   loginID.Value,
						UserID:    user.ID,
					},
				})
			}
		}
	}
	return
}

func (p *SignupProcess) SignupWithOAuth(
	oauthAuthInfo sso.AuthInfo,
	onUserDuplicate model.OnUserDuplicate,
) (principal.Principal, error) {
	authInfo := &authinfo.AuthInfo{}
	principal, err := p.oauthSignupUser(oauthAuthInfo, authInfo, onUserDuplicate)
	if err != nil {
		return nil, err
	}

	userProfile, err := p.UserProfileStore.CreateUserProfile(authInfo.ID, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	user := model.NewUser(*authInfo, userProfile)
	identity := model.NewIdentity(principal)

	err = p.HookProvider.DispatchEvent(
		event.UserCreateEvent{
			User:       user,
			Identities: []model.Identity{identity},
		},
		&user,
	)
	if err != nil {
		return nil, err
	}

	if p.WelcomeEmailConfiguration.Enabled && oauthAuthInfo.ProviderUserInfo.Email != "" {
		p.TaskQueue.Enqueue(async.TaskSpec{
			Name: task.WelcomeEmailSendTaskName,
			Param: task.WelcomeEmailSendTaskParam{
				URLPrefix: p.URLPrefixProvider.Value(),
				Email:     oauthAuthInfo.ProviderUserInfo.Email,
				User:      user,
			},
		})
	}

	return principal, nil
}

func (p *SignupProcess) oauthSignupUser(oauthAuthInfo sso.AuthInfo, authInfo *authinfo.AuthInfo, onUserDuplicate model.OnUserDuplicate) (oauthPrincipal *oauth.Principal, err error) {
	now := p.TimeProvider.NowUTC()
	createFunc := func() {
		*authInfo = authinfo.NewAuthInfo()
		authInfo.LastLoginAt = &now

		// Create AuthInfo
		if err = p.AuthInfoStore.CreateAuth(authInfo); err != nil {
			return
		}

		oauthPrincipal, err = p.createOAuthPrincipal(authInfo.ID, now, oauthAuthInfo)
		return
	}

	// Case: OAuth principal was not found
	// We need to consider all principals
	realm := password.DefaultRealm
	principals, err := p.findExistingPrincipalsWithEmail(oauthAuthInfo.ProviderUserInfo.Email, realm)
	if err != nil {
		return
	}
	userIDs := principalsToUserIDs(principals)

	// Case: OAuth principal was not found and no other principals were not found
	// => Simple create case
	if len(userIDs) <= 0 {
		createFunc()
		return
	}

	// Case: OAuth principal was not found and some principals were found
	// => Complex case
	switch onUserDuplicate {
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
		// Need to associate the provider to the existing user
		return nil, &oAuthRequireMergeError{UserID: userIDs[0]}
	}

	return
}
func (p *SignupProcess) LinkWithOAuth(
	oauthAuthInfo sso.AuthInfo,
	userID string,
) (principal.Principal, error) {
	_, err := p.OAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
		ProviderType:   string(oauthAuthInfo.ProviderConfig.Type),
		ProviderKeys:   oauth.ProviderKeysFromProviderConfig(oauthAuthInfo.ProviderConfig),
		ProviderUserID: oauthAuthInfo.ProviderUserInfo.ID,
	})
	if err == nil {
		return nil, sso.NewSSOFailed(sso.AlreadyLinked, "user is already linked to this provider")
	}

	if !errors.Is(err, principal.ErrNotFound) {
		return nil, err
	}

	authInfo := &authinfo.AuthInfo{}
	err = p.AuthInfoStore.GetAuth(userID, authInfo)
	if err != nil {
		return nil, err
	}

	now := p.TimeProvider.NowUTC()

	var principal *oauth.Principal
	principal, err = p.createOAuthPrincipal(authInfo.ID, now, oauthAuthInfo)
	if err != nil {
		return nil, err
	}

	var userProfile userprofile.UserProfile
	userProfile, err = p.UserProfileStore.GetUserProfile(authInfo.ID)
	if err != nil {
		return nil, err
	}

	user := model.NewUser(*authInfo, userProfile)
	identity := model.NewIdentity(principal)
	err = p.HookProvider.DispatchEvent(
		event.IdentityCreateEvent{
			User:     user,
			Identity: identity,
		},
		&user,
	)
	if err != nil {
		return nil, err
	}

	return principal, nil
}

func (p *SignupProcess) createOAuthPrincipal(userID string, now time.Time, oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	providerKeys := oauth.ProviderKeysFromProviderConfig(oauthAuthInfo.ProviderConfig)
	principal := oauth.NewPrincipal(providerKeys)
	principal.UserID = userID
	principal.ProviderType = string(oauthAuthInfo.ProviderConfig.Type)
	principal.ProviderUserID = oauthAuthInfo.ProviderUserInfo.ID
	principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
	principal.UserProfile = oauthAuthInfo.ProviderRawProfile
	principal.ClaimsValue = oauthAuthInfo.ProviderUserInfo.ClaimsValue()
	principal.CreatedAt = &now
	principal.UpdatedAt = &now
	err := p.OAuthProvider.CreatePrincipal(principal)
	return principal, err
}

func (p *SignupProcess) findExistingPrincipalsWithEmail(email string, mergeRealm string) ([]principal.Principal, error) {
	if email == "" {
		return nil, nil
	}
	principals, err := p.IdentityProvider.ListPrincipalsByClaim("email", email)
	if err != nil {
		return nil, err
	}
	var filteredPrincipals []principal.Principal
	for _, p := range principals {
		if passwordPrincipal, ok := p.(*password.Principal); ok && passwordPrincipal.Realm != mergeRealm {
			continue
		}
		filteredPrincipals = append(filteredPrincipals, p)
	}
	return filteredPrincipals, nil
}
