package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

const (
	googleAuthorizationURL string = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL         string = "https://www.googleapis.com/oauth2/v4/token"
	googleUserInfoURL      string = "https://www.googleapis.com/oauth2/v1/userinfo"
)

type GoogleImpl struct {
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *GoogleImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.ProviderConfig.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	params.Options["access_type"] = "offline"
	params.Options["prompt"] = "select_account"
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		options:        params.Options,
		state:          NewState(params),
		baseURL:        googleAuthorizationURL,
	}
	return authURL(p)
}

func (f *GoogleImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *GoogleImpl) GetAuthInfo(code string, scope string, encodedState string) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		code:           code,
		encodedState:   encodedState,
		accessTokenURL: googleTokenURL,
		userProfileURL: googleUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfo()
}

func (f *GoogleImpl) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: googleTokenURL,
		userProfileURL: googleUserInfoURL,
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ Provider = &GoogleImpl{}
)
