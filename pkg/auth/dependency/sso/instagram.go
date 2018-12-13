package sso

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type InstagramImpl struct {
	Setting Setting
	Config  Config
}

func (f *InstagramImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.Config.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	encodedState, err := EncodeState(f.Setting.StateJWTSecret, NewState(params))
	if err != nil {
		return "", err
	}
	v := url.Values{}
	v.Set("response_type", "code")
	v.Add("client_id", f.Config.ClientID)
	v.Add("redirect_uri", RedirectURI(f.Setting.URLPrefix, f.Config.Name))
	for k, o := range params.Options {
		v.Add(k, fmt.Sprintf("%v", o))
	}
	v.Add("scope", strings.Join(GetScope(params.Scope, f.Config.Scope), " "))
	// Instagram non-compliance fix
	// if we don't put state as the last parameter
	// instagram will convert the state value to lower case
	// when redirecting user to login page if user has not logged in before
	v.Add("state", encodedState)
	return BaseURL(f.Config.Name) + "?" + v.Encode(), nil
}

func (f *InstagramImpl) GetAuthInfo(code string, scope Scope, encodedState string) (authInfo AuthInfo, err error) {
	h := authHandler{
		providerName:   f.Config.Name,
		clientID:       f.Config.ClientID,
		clientSecret:   f.Config.ClientSecret,
		urlPrefix:      f.Setting.URLPrefix,
		code:           code,
		scope:          scope,
		stateJWTSecret: f.Setting.StateJWTSecret,
		encodedState:   encodedState,
		accessTokenURL: AccessTokenURL(f.Config.Name),
		userProfileURL: UserProfileURL(f.Config.Name),
		processUserID:  f.processUserID,
	}
	return h.getAuthInfo()
}

func (f *InstagramImpl) processUserID(userProfile map[string]interface{}) string {
	// Check GET /users/self response
	// https://www.instagram.com/developer/endpoints/users/
	data, ok := userProfile["data"].(map[string]interface{})
	if !ok {
		return ""
	}
	id, ok := data["id"].(string)
	if !ok {
		return ""
	}
	return id
}

func (f *InstagramImpl) processAuthData(userProfile map[string]interface{}) (authData map[string]interface{}) {
	// Check GET /users/self response
	// https://www.instagram.com/developer/endpoints/users/
	authData = make(map[string]interface{})
	data, ok := userProfile["data"].(map[string]interface{})
	if !ok {
		return
	}
	email, ok := data["email"].(string)
	if ok {
		authData["email"] = email
	}
	return
}
