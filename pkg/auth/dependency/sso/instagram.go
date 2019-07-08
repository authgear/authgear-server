package sso

import (
	"fmt"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type InstagramImpl struct {
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

type instagramAuthInfoProcessor struct {
	defaultAuthInfoProcessor
}

func newInstagramAuthInfoProcessor() instagramAuthInfoProcessor {
	return instagramAuthInfoProcessor{}
}

func (f *InstagramImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.ProviderConfig.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	encodedState, err := EncodeState(f.OAuthConfig.StateJWTSecret, NewState(params))
	if err != nil {
		return "", err
	}
	v := url.Values{}
	v.Set("response_type", "code")
	v.Add("client_id", f.ProviderConfig.ClientID)
	v.Add("redirect_uri", RedirectURI(f.OAuthConfig, f.ProviderConfig))
	for k, o := range params.Options {
		v.Add(k, fmt.Sprintf("%v", o))
	}
	v.Add("scope", f.ProviderConfig.Scope)
	// Instagram non-compliance fix
	// if we don't put state as the last parameter
	// instagram will convert the state value to lower case
	// when redirecting user to login page if user has not logged in before
	v.Add("state", encodedState)
	return BaseURL(f.ProviderConfig) + "?" + v.Encode(), nil
}

func (f *InstagramImpl) GetAuthInfo(code string, scope string, encodedState string) (authInfo AuthInfo, err error) {
	p := newInstagramAuthInfoProcessor()
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		code:           code,
		encodedState:   encodedState,
		accessTokenURL: AccessTokenURL(f.ProviderConfig),
		userProfileURL: UserProfileURL(f.ProviderConfig),
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
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: AccessTokenURL(f.ProviderConfig),
		userProfileURL: UserProfileURL(f.ProviderConfig),
		processor:      p,
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}
