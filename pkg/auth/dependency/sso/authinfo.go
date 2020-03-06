package sso

import (
	"crypto/subtle"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

// AuthInfo contains auth info from HandleAuthzResp
type AuthInfo struct {
	ProviderConfig          config.OAuthProviderConfiguration
	ProviderRawProfile      map[string]interface{}
	ProviderAccessTokenResp interface{}
	ProviderUserInfo        ProviderUserInfo
}

type ProviderUserInfo struct {
	ID string
	// Email is normalized.
	Email string
}

func (i ProviderUserInfo) ClaimsValue() map[string]interface{} {
	claimsValue := map[string]interface{}{}
	if i.Email != "" {
		claimsValue["email"] = i.Email
	}
	return claimsValue
}

type OAuthAuthorizationResponse struct {
	Code  string
	State string
	Scope string
	// Nonce is required when the provider supports OpenID connect or OAuth Authorization Code Flow.
	// The implementation is based on the suggestion in the spec.
	// See https://openid.net/specs/openid-connect-core-1_0.html#NonceNotes
	//
	// The nonce is a cryptographically random string.
	// The nonce is stored in the session cookie when auth URL is called.
	// The nonce is hashed with SHA256.
	// The hashed nonce is given to the OIDC provider
	// The hashed nonce is stored in the state.
	// The callback endpoint expect the user agent to include the nonce in the session cookie.
	// The nonce in session cookie will be validated against the hashed nonce in the ID token.
	// The nonce in session cookie will be validated against the hashed nonce in the state.
	Nonce string
}

type getAuthInfoRequest struct {
	urlPrefix       *url.URL
	oauthConfig     *config.OAuthConfiguration
	providerConfig  config.OAuthProviderConfiguration
	accessTokenURL  string
	userProfileURL  string
	userInfoDecoder UserInfoDecoder
}

func (h getAuthInfoRequest) getAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	if subtle.ConstantTimeCompare([]byte(state.Nonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = errors.WithSecondaryError(
			NewSSOFailed(SSOUnauthorized, "unexpected authorization response"),
			errors.New("invalid nonce"),
		)
		return
	}

	// compare nonce
	authInfo = AuthInfo{
		ProviderConfig: h.providerConfig,
	}

	accessTokenResp, err := fetchAccessTokenResp(
		r.Code,
		h.accessTokenURL,
		h.urlPrefix,
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
	providerUserInfo, err := h.userInfoDecoder.DecodeUserInfo(h.providerConfig.Type, userProfile)
	if err != nil {
		return
	}
	authInfo.ProviderUserInfo = *providerUserInfo

	return
}
