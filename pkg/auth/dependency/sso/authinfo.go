package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type getAuthInfoRequest struct {
	oauthConfig    config.OAuthConfiguration
	providerConfig config.OAuthProviderConfiguration
	code           string
	accessTokenURL string
	userProfileURL string
	processor      AuthInfoProcessor
}

func (h getAuthInfoRequest) getAuthInfo() (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderConfig: h.providerConfig,
	}

	accessTokenResp, err := fetchAccessTokenResp(
		h.code,
		h.accessTokenURL,
		h.oauthConfig,
		h.providerConfig,
	)
	if err != nil {
		return
	}

	err = accessTokenResp.Validate()
	if err != nil {
		return
	}
	authInfo.ProviderAccessTokenResp = accessTokenResp

	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

func (h getAuthInfoRequest) getAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderConfig: h.providerConfig,
		// validated accessTokenResp
		ProviderAccessTokenResp: accessTokenResp,
	}

	userProfile, err := fetchUserProfile(accessTokenResp, h.userProfileURL)
	if err != nil {
		return
	}
	authInfo.ProviderRawProfile = userProfile
	authInfo.ProviderUserInfo = h.processor.DecodeUserInfo(userProfile)

	return
}
