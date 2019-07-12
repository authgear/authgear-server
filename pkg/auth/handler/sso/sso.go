package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	signUpHandler "github.com/skygeario/skygear-server/pkg/auth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type respHandler struct {
	TokenStore           authtoken.Store
	AuthInfoStore        authinfo.Store
	OAuthAuthProvider    oauth.Provider
	PasswordAuthProvider password.Provider
	IdentityProvider     principal.IdentityProvider
	UserProfileStore     userprofile.Store
	UserID               string
}

func (h respHandler) loginActionResp(oauthAuthInfo sso.AuthInfo) (resp interface{}, err error) {
	// action => login
	var info authinfo.AuthInfo
	createNewUser, principal, err := h.handleLogin(oauthAuthInfo, &info)
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
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
		return
	}

	// Create auth token
	var token authtoken.Token
	token, err = h.TokenStore.NewToken(info.ID, principal.ID)
	if err != nil {
		panic(err)
	}
	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	user := model.NewUser(info, userProfile)
	identity := model.NewIdentity(h.IdentityProvider, principal)
	resp = model.NewAuthResponse(user, identity, token.AccessToken)

	// Populate the activity time to user
	now := timeNow()
	info.LastLoginAt = &now
	info.LastSeenAt = &now
	if err = h.AuthInfoStore.UpdateAuth(&info); err != nil {
		err = skyerr.MakeError(err)
		return
	}

	return
}

func (h respHandler) linkActionResp(oauthAuthInfo sso.AuthInfo) (resp interface{}, err error) {
	// action => link
	// check if provider user is already linked
	_, err = h.OAuthAuthProvider.GetPrincipalByProviderUserID(oauthAuthInfo.ProviderConfig.ID, oauthAuthInfo.ProviderUserInfo.ID)
	if err == nil {
		err = skyerr.NewError(skyerr.InvalidArgument, "user linked to the provider already")
		return resp, err
	}

	if err != skydb.ErrUserNotFound {
		// some other error
		return resp, err
	}

	// check if user is already linked
	_, err = h.OAuthAuthProvider.GetPrincipalByUserID(oauthAuthInfo.ProviderConfig.ID, h.UserID)
	if err == nil {
		err = skyerr.NewError(skyerr.InvalidArgument, "provider account already linked with existing user")
		return resp, err
	}

	if err != skydb.ErrUserNotFound {
		// some other error
		return resp, err
	}

	var info authinfo.AuthInfo
	if err = h.AuthInfoStore.GetAuth(h.UserID, &info); err != nil {
		err = skyerr.NewError(skyerr.ResourceNotFound, "user not found")
		return resp, err
	}

	_, err = h.createPrincipalByOAuthInfo(info.ID, oauthAuthInfo)
	if err != nil {
		return resp, err
	}
	resp = map[string]string{}
	return
}

func (h respHandler) handleLogin(
	oauthAuthInfo sso.AuthInfo,
	info *authinfo.AuthInfo,
) (createNewUser bool, oauthPrincipal *oauth.Principal, err error) {
	oauthPrincipal, err = h.findExistingOAuthPrincipal(oauthAuthInfo)
	if err != nil {
		return
	}

	now := timeNow()

	// Two func that closes over the arguments and the return value
	// and need to be reused.

	// populateInfo sets the argument info to non-nil value
	populateInfo := func(userID string) {
		if e := h.AuthInfoStore.GetAuth(userID, info); e != nil {
			if e == skydb.ErrUserNotFound {
				err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
				return
			}
			err = skyerr.MakeError(e)
			return
		}
	}

	// createFunc creates a new user.
	createFunc := func() {
		createNewUser = true
		// if there is no existed user
		// signup a new user
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

		oauthPrincipal, err = h.createPrincipalByOAuthInfo(info.ID, oauthAuthInfo)
		if err != nil {
			return
		}
	}

	// Case: OAuth principal was found
	// => Simple update case
	// We do not need to consider password principal
	if oauthPrincipal != nil {
		oauthPrincipal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
		oauthPrincipal.UserProfile = oauthAuthInfo.ProviderRawProfile
		oauthPrincipal.UpdatedAt = &now
		if err = h.OAuthAuthProvider.UpdatePrincipal(oauthPrincipal); err != nil {
			err = skyerr.MakeError(err)
			return
		}
		populateInfo(oauthPrincipal.UserID)
		// Always return here because we are done with this case.
		return
	}

	// Case: OAuth principal was not found
	// We need to consider password principal
	passwordPrincipal, err := h.findExistingPasswordPrincipal(oauthAuthInfo)
	if err != nil {
		return
	}

	// Case: OAuth principal was not found and Password principal was not found
	// => Simple create case
	if passwordPrincipal == nil {
		createFunc()
		return
	}

	// Case: OAuth principal was not found and Password principal was found
	// => Complex case
	switch oauthAuthInfo.State.OnUserDuplicate {
	case sso.OnUserDuplicateAbort:
		err = skyerr.NewError(skyerr.Duplicated, "Aborted due to duplicate user")
	case sso.OnUserDuplicateCreate:
		createFunc()
	case sso.OnUserDuplicateMerge:
		// Associate the provider to the existing user
		oauthPrincipal, err = h.createPrincipalByOAuthInfo(
			passwordPrincipal.UserID,
			oauthAuthInfo,
		)
		if err != nil {
			return
		}
		populateInfo(passwordPrincipal.UserID)
	}

	return
}

func (h respHandler) findExistingOAuthPrincipal(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	// Find oauth principal from by (provider_id, provider_user_id)
	principal, err := h.OAuthAuthProvider.GetPrincipalByProviderUserID(oauthAuthInfo.ProviderConfig.ID, oauthAuthInfo.ProviderUserInfo.ID)
	if err == skydb.ErrUserNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return principal, nil
}

func (h respHandler) findExistingPasswordPrincipal(oauthAuthInfo sso.AuthInfo) (*password.Principal, error) {
	// Find password principal by provider primary email
	email := oauthAuthInfo.ProviderUserInfo.Email
	if email == "" {
		return nil, nil
	}
	passwordPrincipal := password.Principal{}
	err := h.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("", email, oauthAuthInfo.State.MergeRealm, &passwordPrincipal)
	if err == skydb.ErrUserNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !h.PasswordAuthProvider.CheckLoginIDKeyType(passwordPrincipal.LoginIDKey, metadata.Email) {
		return nil, nil
	}
	return &passwordPrincipal, nil
}

func (h respHandler) createPrincipalByOAuthInfo(userID string, oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	now := timeNow()
	principal := oauth.NewPrincipal()
	principal.UserID = userID
	principal.ProviderName = oauthAuthInfo.ProviderConfig.ID
	principal.ProviderUserID = oauthAuthInfo.ProviderUserInfo.ID
	principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
	principal.UserProfile = oauthAuthInfo.ProviderRawProfile
	principal.CreatedAt = &now
	principal.UpdatedAt = &now
	err := h.OAuthAuthProvider.CreatePrincipal(principal)
	return &principal, err
}
