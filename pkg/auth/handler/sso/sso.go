package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	signUpHandler "github.com/skygeario/skygear-server/pkg/auth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type respHandler struct {
	TokenStore           authtoken.Store
	AuthInfoStore        authinfo.Store
	OAuthAuthProvider    oauth.Provider
	PasswordAuthProvider password.Provider
	UserProfileStore     userprofile.Store
	UserID               string
}

func (h respHandler) loginActionResp(oauthAuthInfo sso.AuthInfo) (resp interface{}, err error) {
	// action => login
	var info authinfo.AuthInfo
	createNewUser, err := h.handleLogin(oauthAuthInfo, &info)
	if err != nil {
		return
	}

	// Create or update user profile
	var userProfile userprofile.UserProfile
	// oauthAuthInfo.ProviderUserProfile may contains attributes like "id",
	// and it is not allowed to use it in SDK.
	// so here we will save authData as providerUserProfile
	data := make(map[string]interface{})
	providerUserProfile := oauthAuthInfo.ProviderAuthData
	// convert from map[string]string(sso.AuthInfo.ProviderAuthData) to map[string]interface(userprofile.Data)
	for k, v := range providerUserProfile {
		data[k] = v
	}
	if createNewUser {
		userProfile, err = h.UserProfileStore.CreateUserProfile(info.ID, data)
	} else {
		userProfile, err = h.UserProfileStore.UpdateUserProfile(info.ID, &info, data)
	}
	if err != nil {
		// TODO:
		// return proper error
		err = skyerr.NewError(skyerr.UnexpectedError, "Unable to save user profile")
		return
	}

	// Create auth token
	var token authtoken.Token
	token, err = h.TokenStore.NewToken(info.ID)
	if err != nil {
		panic(err)
	}
	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	respFactory := response.AuthResponseFactory{}
	resp = respFactory.NewAuthResponse(info, userProfile, token.AccessToken)

	// Populate the activity time to user
	now := timeNow()
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
	_, err = h.OAuthAuthProvider.GetPrincipalByProviderUserID(oauthAuthInfo.ProviderName, oauthAuthInfo.ProviderUserID)
	if err == nil {
		err = skyerr.NewError(skyerr.InvalidArgument, "user linked to the provider already")
		return resp, err
	}

	if err != skydb.ErrUserNotFound {
		// some other error
		return resp, err
	}

	// check if user is already linked
	_, err = h.OAuthAuthProvider.GetPrincipalByUserID(oauthAuthInfo.ProviderName, h.UserID)
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
	resp = "OK"
	return
}

func (h respHandler) handleLogin(
	oauthAuthInfo sso.AuthInfo,
	info *authinfo.AuthInfo,
) (createNewUser bool, err error) {
	principal, err := h.findPrincipal(oauthAuthInfo)
	if err != nil {
		return
	}

	now := timeNow()
	if principal == nil {
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

		_, err = h.createPrincipalByOAuthInfo(info.ID, oauthAuthInfo)
		if err != nil {
			return
		}

		err = h.createEmptyPasswordPrincipal(info.ID, oauthAuthInfo)
		if err == skydb.ErrUserDuplicated {
			err = signUpHandler.ErrUserDuplicated
		}
	} else {
		principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
		principal.UserProfile = oauthAuthInfo.ProviderUserProfile
		principal.UpdatedAt = &now

		if err = h.OAuthAuthProvider.UpdatePrincipal(principal); err != nil {
			err = skyerr.MakeError(err)
			return
		}

		if e := h.AuthInfoStore.GetAuth(principal.UserID, info); e != nil {
			if err == skydb.ErrUserNotFound {
				err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
				return
			}
			err = skyerr.NewError(skyerr.ResourceNotFound, "User not found")
			return
		}
	}
	return
}

func (h respHandler) findPrincipal(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	// find oauth principal from principal_oauth
	principal, err := h.OAuthAuthProvider.GetPrincipalByProviderUserID(oauthAuthInfo.ProviderName, oauthAuthInfo.ProviderUserID)
	if err != nil {
		if err != skydb.ErrUserNotFound {
			return nil, err
		}
	} else {
		return principal, nil
	}

	// if oauth principal doesn't exist, try to link existed password principal
	if valid := h.PasswordAuthProvider.IsLoginIDValid(oauthAuthInfo.ProviderAuthData); valid {
		// provider authData matches app's loginIDsKeyWhitelist,
		// then it starts auto-link procedure.
		//
		// for example, if oauthAuthInfo.ProviderAuthData is {"email", "john.doe@example.com"},
		// it will be a valid authData if loginIDsKeyWhitelist is [](empty), ["username", "email"] or ["email"]
		// so, the oauthAuthInfo.ProviderAuthDat can be used as a password principal authData
		return h.authLinkUser(oauthAuthInfo)
	}

	return nil, nil
}

func (h respHandler) authLinkUser(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	passwordPrincipal := password.Principal{}
	var e error
	if email, ok := oauthAuthInfo.ProviderAuthData["email"]; ok {
		e = h.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("email", email, password.DefaultRealm, &passwordPrincipal)
	}
	if e == nil {
		userID := passwordPrincipal.UserID
		// link password principal to oauth principal
		oauthPrincipal, err := h.createPrincipalByOAuthInfo(userID, oauthAuthInfo)
		if err != nil {
			return nil, err
		}
		return &oauthPrincipal, nil
	} else if e != skydb.ErrUserNotFound {
		return nil, e
	}

	return nil, nil
}

func (h respHandler) createPrincipalByOAuthInfo(userID string, oauthAuthInfo sso.AuthInfo) (oauth.Principal, error) {
	now := timeNow()
	principal := oauth.NewPrincipal()
	principal.UserID = userID
	principal.ProviderName = oauthAuthInfo.ProviderName
	principal.ProviderUserID = oauthAuthInfo.ProviderUserID
	principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
	principal.UserProfile = oauthAuthInfo.ProviderUserProfile
	principal.CreatedAt = &now
	principal.UpdatedAt = &now
	err := h.OAuthAuthProvider.CreatePrincipal(principal)
	return principal, err
}

func (h respHandler) createEmptyPasswordPrincipal(userID string, oauthAuthInfo sso.AuthInfo) error {
	if valid := h.PasswordAuthProvider.IsLoginIDValid(oauthAuthInfo.ProviderAuthData); valid {
		// if ProviderAuthData matches loginIDsKeyWhitelist, and it can't be link with current account,
		// we also creates an empty password principal for later the user can set password to it
		return h.PasswordAuthProvider.CreatePrincipalsByLoginID(userID, "", oauthAuthInfo.ProviderAuthData)
	}

	return nil
}
