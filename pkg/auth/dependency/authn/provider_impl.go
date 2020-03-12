package authn

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

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
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type ProviderImpl struct {
	Logger                        *logrus.Entry
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
}

var _ Provider = &ProviderImpl{}

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

func (p *ProviderImpl) CreateUserWithLoginIDs(
	loginIDs []loginid.LoginID,
	plainPassword string,
	metadata map[string]interface{},
	onUserDuplicate model.OnUserDuplicate,
) (authInfo *authinfo.AuthInfo, userProfile *userprofile.UserProfile, firstPrincipal principal.Principal, tasks []async.TaskSpec, err error) {
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
	authInfo = &authInfoS
	authInfo.LastLoginAt = &now

	err = p.AuthInfoStore.CreateAuth(authInfo)
	if err != nil {
		return
	}

	userProfileS, err := p.UserProfileStore.CreateUserProfile(authInfo.ID, metadata)
	if err != nil {
		return
	}
	userProfile = &userProfileS

	principals, err := p.createPrincipalsWithProposedLoginIDs(authInfo.ID, plainPassword, loginIDs)
	if err != nil {
		return
	}
	firstPrincipal = principals[0]

	user := model.NewUser(*authInfo, *userProfile)
	identities := []model.Identity{}
	for _, principal := range principals {
		identity := model.NewIdentity(p.IdentityProvider, principal)
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
		tasks = append(tasks, p.generateSendWelcomeEmailTasks(user, loginIDs)...)
	}

	if p.UserVerificationConfiguration.AutoSendOnSignup {
		tasks = append(tasks, p.generateSendVerificationCodeTasks(user, loginIDs)...)
	}

	return
}

func (p *ProviderImpl) validateCreateUserWithLoginIDs(loginIDs []loginid.LoginID, onUserDuplicate model.OnUserDuplicate) (err error) {
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

func (p *ProviderImpl) findExistingPrincipalsWithProposedLoginIDs(loginIDs []loginid.LoginID) (principals []principal.Principal, err error) {
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

func (p *ProviderImpl) createPrincipalsWithProposedLoginIDs(userID string, plainPassword string, loginIDs []loginid.LoginID) (principals []principal.Principal, err error) {
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

func (p *ProviderImpl) generateSendWelcomeEmailTasks(user model.User, loginIDs []loginid.LoginID) (tasks []async.TaskSpec) {
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
		tasks = append(tasks, async.TaskSpec{
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

func (p *ProviderImpl) generateSendVerificationCodeTasks(user model.User, loginIDs []loginid.LoginID) (tasks []async.TaskSpec) {
	for _, loginID := range loginIDs {
		for _, keyConfig := range p.UserVerificationConfiguration.LoginIDKeys {
			if keyConfig.Key == loginID.Key {
				tasks = append(tasks, async.TaskSpec{
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

func (p *ProviderImpl) AuthenticateWithLoginID(loginID loginid.LoginID, plainPassword string) (authInfo *authinfo.AuthInfo, prin principal.Principal, err error) {
	var passwordPrincipal password.Principal
	realm := password.DefaultRealm
	err = p.PasswordProvider.GetPrincipalByLoginIDWithRealm(loginID.Key, loginID.Value, realm, &passwordPrincipal)
	if err != nil {
		if errors.Is(err, principal.ErrNotFound) {
			err = password.ErrInvalidCredentials
		}
		if errors.Is(err, principal.ErrMultipleResultsFound) {
			p.Logger.WithError(err).Warn("multiple results found for password principal query")
			err = password.ErrInvalidCredentials
		}
		return
	}

	err = passwordPrincipal.VerifyPassword(plainPassword)
	if err != nil {
		return
	}

	if err := p.PasswordProvider.MigratePassword(&passwordPrincipal, plainPassword); err != nil {
		p.Logger.WithError(err).Error("failed to migrate password")
	}

	var authInfoS authinfo.AuthInfo
	err = p.AuthInfoStore.GetAuth(passwordPrincipal.UserID, &authInfoS)
	if err != nil {
		return
	}

	authInfo = &authInfoS
	prin = &passwordPrincipal
	return
}

func (p *ProviderImpl) AuthenticateWithOAuth(oauthAuthInfo sso.AuthInfo, codeChallenge string, loginState sso.LoginState) (code *sso.SkygearAuthorizationCode, tasks []async.TaskSpec, err error) {
	var authInfo authinfo.AuthInfo
	createNewUser, principal, err := p.oauthLogin(oauthAuthInfo, &authInfo, loginState)
	if err != nil {
		return
	}

	var userProfile userprofile.UserProfile
	emptyProfile := map[string]interface{}{}
	if createNewUser {
		userProfile, err = p.UserProfileStore.CreateUserProfile(authInfo.ID, emptyProfile)
	} else {
		userProfile, err = p.UserProfileStore.GetUserProfile(authInfo.ID)
	}
	if err != nil {
		return
	}

	user := model.NewUser(authInfo, userProfile)
	identity := model.NewIdentity(p.IdentityProvider, principal)

	if createNewUser {
		err = p.HookProvider.DispatchEvent(
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

	code = &sso.SkygearAuthorizationCode{
		Action:              "login",
		CodeChallenge:       codeChallenge,
		UserID:              user.ID,
		PrincipalID:         principal.ID,
		SessionCreateReason: string(sessionCreateReason),
	}

	if createNewUser && p.WelcomeEmailConfiguration.Enabled && oauthAuthInfo.ProviderUserInfo.Email != "" {
		tasks = append(tasks, async.TaskSpec{
			Name: task.WelcomeEmailSendTaskName,
			Param: task.WelcomeEmailSendTaskParam{
				URLPrefix: p.URLPrefixProvider.Value(),
				Email:     oauthAuthInfo.ProviderUserInfo.Email,
				User:      user,
			},
		})
	}

	return
}

func (p *ProviderImpl) oauthLogin(oauthAuthInfo sso.AuthInfo, authInfo *authinfo.AuthInfo, loginState sso.LoginState) (createNewUser bool, oauthPrincipal *oauth.Principal, err error) {
	oauthPrincipal, err = p.findExistingOAuthPrincipal(oauthAuthInfo)
	if err != nil && !errors.Is(err, principal.ErrNotFound) {
		return
	}

	now := p.TimeProvider.NowUTC()

	// Two func that closes over the arguments and the return value
	// and need to be reused.

	// createFunc creates a new user.
	createFunc := func() {
		createNewUser = true
		// if there is no existed user
		// signup a new user
		*authInfo = authinfo.NewAuthInfo()
		authInfo.LastLoginAt = &now

		// Create AuthInfo
		if err = p.AuthInfoStore.CreateAuth(authInfo); err != nil {
			return
		}

		oauthPrincipal, err = p.createOAuthPrincipal(authInfo.ID, now, oauthAuthInfo)
		if err != nil {
			return
		}
	}

	// Case: OAuth principal was found
	// => Simple update case
	// We do not need to consider other principals
	if err == nil {
		oauthPrincipal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
		oauthPrincipal.UserProfile = oauthAuthInfo.ProviderRawProfile
		oauthPrincipal.ClaimsValue = oauthAuthInfo.ProviderUserInfo.ClaimsValue()
		oauthPrincipal.UpdatedAt = &now
		if err = p.OAuthProvider.UpdatePrincipal(oauthPrincipal); err != nil {
			return
		}
		err = p.AuthInfoStore.GetAuth(oauthPrincipal.UserID, authInfo)
		// Always return here because we are done with this case.
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
	switch loginState.OnUserDuplicate {
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
		oauthPrincipal, err = p.createOAuthPrincipal(
			userID,
			now,
			oauthAuthInfo,
		)
		if err != nil {
			return
		}
		err = p.AuthInfoStore.GetAuth(userID, authInfo)
	}

	return
}

func (p *ProviderImpl) findExistingOAuthPrincipal(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	// Find oauth principal from by (provider_id, provider_user_id)
	return p.OAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
		ProviderType:   string(oauthAuthInfo.ProviderConfig.Type),
		ProviderKeys:   oauth.ProviderKeysFromProviderConfig(oauthAuthInfo.ProviderConfig),
		ProviderUserID: oauthAuthInfo.ProviderUserInfo.ID,
	})
}

func (p *ProviderImpl) findExistingPrincipalsWithEmail(email string, mergeRealm string) ([]principal.Principal, error) {
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

func (p *ProviderImpl) createOAuthPrincipal(userID string, now time.Time, oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
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

func (p *ProviderImpl) LinkOAuth(oauthAuthInfo sso.AuthInfo, codeChallenge string, linkState sso.LinkState) (code *sso.SkygearAuthorizationCode, err error) {
	// action => link
	// We only need to check if we can find such principal.
	// If such principal exists, it does not matter whether the principal
	// is associated with the user.
	// We do not allow the same provider user to be associated with an user
	// more than once.
	_, err = p.OAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
		ProviderType:   string(oauthAuthInfo.ProviderConfig.Type),
		ProviderKeys:   oauth.ProviderKeysFromProviderConfig(oauthAuthInfo.ProviderConfig),
		ProviderUserID: oauthAuthInfo.ProviderUserInfo.ID,
	})
	if err == nil {
		err = sso.NewSSOFailed(sso.AlreadyLinked, "user is already linked to this provider")
		return
	}

	if !errors.Is(err, principal.ErrNotFound) {
		return
	}

	var authInfo authinfo.AuthInfo
	err = p.AuthInfoStore.GetAuth(linkState.UserID, &authInfo)
	if err != nil {
		return
	}

	now := p.TimeProvider.NowUTC()

	var principal *oauth.Principal
	principal, err = p.createOAuthPrincipal(authInfo.ID, now, oauthAuthInfo)
	if err != nil {
		return
	}

	var userProfile userprofile.UserProfile
	userProfile, err = p.UserProfileStore.GetUserProfile(authInfo.ID)
	if err != nil {
		return
	}

	user := model.NewUser(authInfo, userProfile)
	identity := model.NewIdentity(p.IdentityProvider, principal)
	err = p.HookProvider.DispatchEvent(
		event.IdentityCreateEvent{
			User:     user,
			Identity: identity,
		},
		&user,
	)
	if err != nil {
		return
	}

	code = &sso.SkygearAuthorizationCode{
		Action:        "link",
		CodeChallenge: codeChallenge,
		UserID:        user.ID,
		PrincipalID:   principal.ID,
	}

	return
}

func (p *ProviderImpl) ExtractAuthorizationCode(code *sso.SkygearAuthorizationCode) (authInfo *authinfo.AuthInfo, userProfile *userprofile.UserProfile, prin principal.Principal, err error) {
	var authInfoS authinfo.AuthInfo
	if err = p.AuthInfoStore.GetAuth(code.UserID, &authInfoS); err != nil {
		return
	}
	authInfo = &authInfoS

	var userProfileS userprofile.UserProfile
	userProfileS, err = p.UserProfileStore.GetUserProfile(authInfo.ID)
	if err != nil {
		return
	}
	userProfile = &userProfileS

	prin, err = p.IdentityProvider.GetPrincipalByID(code.PrincipalID)
	if err != nil {
		return
	}

	return
}
