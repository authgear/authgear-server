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
	Settings             sso.Setting
}

func (h respHandler) loginActionResp(oauthAuthInfo sso.AuthInfo) (resp interface{}, err error) {
	// action => login
	var info authinfo.AuthInfo
	createNewUser, principal, err := h.handleLogin(oauthAuthInfo, &info)
	if err != nil {
		return
	}

	// Create or update user profile
	var userProfile userprofile.UserProfile
	data := oauthAuthInfo.ProviderRawProfile
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
	token, err = h.TokenStore.NewToken(info.ID, principal.ID)
	if err != nil {
		panic(err)
	}
	if err = h.TokenStore.Put(&token); err != nil {
		panic(err)
	}

	user := model.NewUser(info, userProfile, model.NewIdentity(h.IdentityProvider, principal))
	resp = model.NewAuthResponse(user, token.AccessToken)

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
	_, err = h.OAuthAuthProvider.GetPrincipalByProviderUserID(oauthAuthInfo.ProviderName, oauthAuthInfo.ProviderUserInfo.ID)
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
) (createNewUser bool, principal *oauth.Principal, err error) {
	principal, err = h.findPrincipal(oauthAuthInfo)
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

		principal, err = h.createPrincipalByOAuthInfo(info.ID, oauthAuthInfo)
		if err != nil {
			return
		}

	} else {
		principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
		principal.UserProfile = oauthAuthInfo.ProviderRawProfile
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
		info.LastLoginAt = &now
		if e := h.AuthInfoStore.UpdateAuth(info); e != nil {
			err = skyerr.NewError(skyerr.ResourceNotFound, "Unable to update user")
			return
		}
	}
	return
}

func (h respHandler) findPrincipal(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	// find oauth principal from principal_oauth
	principal, err := h.OAuthAuthProvider.GetPrincipalByProviderUserID(oauthAuthInfo.ProviderName, oauthAuthInfo.ProviderUserInfo.ID)
	if err != nil {
		if err != skydb.ErrUserNotFound {
			return nil, err
		}
	} else {
		return principal, nil
	}

	// if oauth principal doesn't exist, try to link existed password principal
	if h.Settings.AutoLinkEnabled && h.PasswordAuthProvider.IsDefaultAllowedRealms() {
		return h.authLinkUser(oauthAuthInfo)
	}

	return nil, nil
}

func (h respHandler) authLinkUser(oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	email := oauthAuthInfo.ProviderUserInfo.Email
	if email == "" {
		return nil, nil
	}

	var err error
	passwordPrincipal := password.Principal{}
	err = h.PasswordAuthProvider.GetPrincipalByLoginIDWithRealm("", email, password.DefaultRealm, &passwordPrincipal)
	if err != nil {
		if err == skydb.ErrUserNotFound {
			return nil, nil
		}
		return nil, err
	}

	if !h.PasswordAuthProvider.CheckLoginIDKeyType(passwordPrincipal.LoginIDKey, metadata.Email) {
		return nil, nil
	}

	userID := passwordPrincipal.UserID
	// link password principal to oauth principal
	oauthPrincipal, err := h.createPrincipalByOAuthInfo(userID, oauthAuthInfo)
	if err != nil {
		return nil, err
	}
	return oauthPrincipal, nil
}

func (h respHandler) createPrincipalByOAuthInfo(userID string, oauthAuthInfo sso.AuthInfo) (*oauth.Principal, error) {
	now := timeNow()
	principal := oauth.NewPrincipal()
	principal.UserID = userID
	principal.ProviderName = oauthAuthInfo.ProviderName
	principal.ProviderUserID = oauthAuthInfo.ProviderUserInfo.ID
	principal.AccessTokenResp = oauthAuthInfo.ProviderAccessTokenResp
	principal.UserProfile = oauthAuthInfo.ProviderRawProfile
	principal.CreatedAt = &now
	principal.UpdatedAt = &now
	err := h.OAuthAuthProvider.CreatePrincipal(principal)
	return &principal, err
}
