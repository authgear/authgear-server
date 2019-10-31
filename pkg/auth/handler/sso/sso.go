package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

// nolint: deadcode
/*
	@ID SSOProviderID
	@Parameter provider_id path
		ID of SSO provider
		@JSONSchema
			{ "type": "string" }
*/
type ssoProviderParameter string

type respHandler struct {
	AuthnSessionProvider authnsession.Provider
	AuthInfoStore        authinfo.Store
	OAuthAuthProvider    oauth.Provider
	IdentityProvider     principal.IdentityProvider
	UserProfileStore     userprofile.Store
	HookProvider         hook.Provider
	TaskQueue            async.Queue
	WelcomeEmailEnabled  bool
	URLPrefix            *url.URL
}

func (h respHandler) loginActionResp(oauthAuthInfo sso.AuthInfo, loginState sso.LoginState) (resp interface{}, err error) {
	// action => login
	var info authinfo.AuthInfo
	createNewUser, principal, err := h.handleLogin(oauthAuthInfo, &info, loginState)
	if err != nil {
		return
	}

	// Create empty user profile or get the existing one
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

	if createNewUser &&
		h.WelcomeEmailEnabled &&
		oauthAuthInfo.ProviderUserInfo.Email != "" &&
		h.TaskQueue != nil {
		h.TaskQueue.Enqueue(task.WelcomeEmailSendTaskName, task.WelcomeEmailSendTaskParam{
			URLPrefix: h.URLPrefix,
			Email:     oauthAuthInfo.ProviderUserInfo.Email,
			User:      user,
		}, nil)
	}

	return
}

func (h respHandler) linkActionResp(oauthAuthInfo sso.AuthInfo, linkState sso.LinkState) (resp interface{}, err error) {
	// action => link
	// We only need to check if we can find such principal.
	// If such principal exists, it does not matter whether the principal
	// is associated with the user.
	// We do not allow the same provider user to be associated with an user
	// more than once.
	_, err = h.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
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

	var info authinfo.AuthInfo
	if err = h.AuthInfoStore.GetAuth(linkState.UserID, &info); err != nil {
		return
	}

	var principal *oauth.Principal
	principal, err = h.createPrincipalByOAuthInfo(info.ID, oauthAuthInfo)
	if err != nil {
		return
	}

	var userProfile userprofile.UserProfile
	userProfile, err = h.UserProfileStore.GetUserProfile(info.ID)
	if err != nil {
		return
	}

	user := model.NewUser(info, userProfile)
	identity := model.NewIdentity(h.IdentityProvider, principal)
	err = h.HookProvider.DispatchEvent(
		event.IdentityCreateEvent{
			User:     user,
			Identity: identity,
		},
		&user,
	)
	if err != nil {
		return
	}

	resp = model.NewAuthResponseWithUser(user)
	return
}

func (h respHandler) handleLogin(
	oauthAuthInfo sso.AuthInfo,
	info *authinfo.AuthInfo,
	loginState sso.LoginState,
) (createNewUser bool, oauthPrincipal *oauth.Principal, err error) {
	oauthPrincipal, err = h.findExistingOAuthPrincipal(oauthAuthInfo)
	if err != nil && !errors.Is(err, principal.ErrNotFound) {
		return
	}

	now := timeNow()

	// Two func that closes over the arguments and the return value
	// and need to be reused.

	// createFunc creates a new user.
	createFunc := func() {
		createNewUser = true
		// if there is no existed user
		// signup a new user
		*info = authinfo.NewAuthInfo()
		info.LastLoginAt = &now

		// Create AuthInfo
		if err = h.AuthInfoStore.CreateAuth(info); err != nil {
			return
		}

		oauthPrincipal, err = h.createPrincipalByOAuthInfo(info.ID, oauthAuthInfo)
		if err != nil {
			return
		}
	}

	// Case: OAuth principal was found
	// => Simple update case
	// We do not need to consider other principals
	if err == nil {
		oauthPrincipal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
		oauthPrincipal.SetRawProfile(oauthAuthInfo.ProviderRawProfile)
		oauthPrincipal.UpdatedAt = &now
		if err = h.OAuthAuthProvider.UpdatePrincipal(oauthPrincipal); err != nil {
			return
		}
		err = h.AuthInfoStore.GetAuth(oauthPrincipal.UserID, info)
		// Always return here because we are done with this case.
		return
	}

	// Case: OAuth principal was not found
	// We need to consider all principals
	principals, err := h.findExistingPrincipals(oauthAuthInfo, loginState.MergeRealm)
	if err != nil {
		return
	}
	userIDs := h.principalsToUserIDs(principals)

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
		oauthPrincipal, err = h.createPrincipalByOAuthInfo(
			userID,
			oauthAuthInfo,
		)
		if err != nil {
			return
		}
		err = h.AuthInfoStore.GetAuth(userID, info)
	}

	return
}

func (h respHandler) findExistingOAuthPrincipal(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	// Find oauth principal from by (provider_id, provider_user_id)
	return h.OAuthAuthProvider.GetPrincipalByProvider(oauth.GetByProviderOptions{
		ProviderType:   string(oauthAuthInfo.ProviderConfig.Type),
		ProviderKeys:   oauth.ProviderKeysFromProviderConfig(oauthAuthInfo.ProviderConfig),
		ProviderUserID: oauthAuthInfo.ProviderUserInfo.ID,
	})
}

func (h respHandler) findExistingPrincipals(oauthAuthInfo sso.AuthInfo, mergeRealm string) ([]principal.Principal, error) {
	email := oauthAuthInfo.ProviderUserInfo.Email
	if email == "" {
		return nil, nil
	}
	principals, err := h.IdentityProvider.ListPrincipalsByClaim("email", oauthAuthInfo.ProviderUserInfo.Email)
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

func (h respHandler) principalsToUserIDs(principals []principal.Principal) []string {
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

func (h respHandler) createPrincipalByOAuthInfo(userID string, oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	now := timeNow()
	providerKeys := oauth.ProviderKeysFromProviderConfig(oauthAuthInfo.ProviderConfig)
	principal := oauth.NewPrincipal(providerKeys)
	principal.UserID = userID
	principal.ProviderType = string(oauthAuthInfo.ProviderConfig.Type)
	principal.ProviderUserID = oauthAuthInfo.ProviderUserInfo.ID
	principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
	principal.SetRawProfile(oauthAuthInfo.ProviderRawProfile)
	principal.CreatedAt = &now
	principal.UpdatedAt = &now
	err := h.OAuthAuthProvider.CreatePrincipal(principal)
	return principal, err
}
