package sso

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type InstagramImpl struct {
	Setting Setting
	Config  Config
}

type instagramAuthInfoProcessor struct {
	defaultAuthInfoProcessor
}

func newInstagramAuthInfoProcessor() instagramAuthInfoProcessor {
	return instagramAuthInfoProcessor{}
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
	p := newInstagramAuthInfoProcessor()
	h := getAuthInfoRequest{
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
		processor:      p,
	}
	return h.getAuthInfo()
}

func (i instagramAuthInfoProcessor) DecodeUserInfo(userProfile map[string]interface{}) (info ProviderUserInfo) {
	// Check GET /users/self response
	// https://www.instagram.com/developer/endpoints/users/
	data, ok := userProfile["data"].(map[string]interface{})
	if !ok {
		return
	}

	info.ID, _ = data["id"].(string)
	info.Email, _ = data["email"].(string)
	return
}

func (f *InstagramImpl) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	p := newInstagramAuthInfoProcessor()
	h := getAuthInfoRequest{
		providerName:   f.Config.Name,
		clientID:       f.Config.ClientID,
		clientSecret:   f.Config.ClientSecret,
		urlPrefix:      f.Setting.URLPrefix,
		stateJWTSecret: f.Setting.StateJWTSecret,
		accessTokenURL: AccessTokenURL(f.Config.Name),
		userProfileURL: UserProfileURL(f.Config.Name),
		processor:      p,
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}
