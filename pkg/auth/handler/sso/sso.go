package sso

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type respHandler struct {
	RoleStore            role.Store
	TokenStore           authtoken.Store
	AuthInfoStore        authinfo.Store
	OAuthAuthProvider    oauth.Provider
	PasswordAuthProvider password.Provider
	UserProfileStore     userprofile.Store
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
	providerUserProfile := oauthAuthInfo.ProviderAuthData
	if createNewUser {
		userProfile, err = h.UserProfileStore.CreateUserProfile(info.ID, &info, providerUserProfile)
	} else {
		userProfile, err = h.UserProfileStore.UpdateUserProfile(info.ID, &info, providerUserProfile)
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

	resp = response.NewAuthResponse(info, userProfile, token.AccessToken)

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
	userID := oauthAuthInfo.State.UserID // skygear userID
	_, err = h.OAuthAuthProvider.GetPrincipalByUserID(userID)
	if err == nil {
		err = skyerr.NewError(skyerr.InvalidArgument, "provider account already linked with existing user")
		return resp, err
	}

	if err != skydb.ErrUserNotFound {
		// some other error
		return resp, err
	}

	var info authinfo.AuthInfo
	if err = h.AuthInfoStore.GetAuth(userID, &info); err != nil {
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
	now := timeNow()

	principal, err := h.OAuthAuthProvider.GetPrincipalByProviderUserID(oauthAuthInfo.ProviderName, oauthAuthInfo.ProviderUserID)
	if err != nil {
		if err != skydb.ErrUserNotFound {
			return
		}
		err = nil
	}

	if valid := h.PasswordAuthProvider.IsAuthDataMatching(oauthAuthInfo.ProviderAuthData); valid {
		// provider authData match app authRecordKeys,
		// start auto-connect procedure
		principal, err = h.authLinkUser(oauthAuthInfo)
		if err != nil {
			return
		}
	}

	if principal == nil {
		createNewUser = true
		// if there is no existed user
		// signup a new user
		*info = authinfo.NewAuthInfo()
		info.LastLoginAt = &now

		// Get default roles
		defaultRoles, e := h.RoleStore.GetDefaultRoles()
		if e != nil {
			err = skyerr.NewError(skyerr.InternalQueryInvalid, "unable to query default roles")
			return
		}

		// Assign default roles
		info.Roles = defaultRoles

		// Create AuthInfo
		if e = h.AuthInfoStore.CreateAuth(info); e != nil {
			if e == skydb.ErrUserDuplicated {
				err = skyerr.NewError(skyerr.Duplicated, "user duplicated")
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

func (h respHandler) authLinkUser(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	principal := password.Principal{}
	e := h.PasswordAuthProvider.GetPrincipalByAuthData(oauthAuthInfo.ProviderAuthData, &principal)
	if e == nil {
		userID := principal.UserID
		// link user
		principal, err := h.createPrincipalByOAuthInfo(userID, oauthAuthInfo)
		if err != nil {
			return nil, err
		}
		return &principal, nil
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
	if valid := h.PasswordAuthProvider.IsAuthDataMatching(oauthAuthInfo.ProviderAuthData); valid {
		// if ProviderAuthData mastch authRecordKeys, also creates an empty password principal
		// and later the user can set password to it
		return h.PasswordAuthProvider.CreatePrincipalsByAuthData(userID, "", oauthAuthInfo.ProviderAuthData)
	}

	return nil
}
