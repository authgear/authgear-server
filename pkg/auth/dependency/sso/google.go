package sso

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

const (
	googleOIDCDiscoveryDocumentURL string = "https://accounts.google.com/.well-known/openid-configuration"
	// nolint: gosec
	googleTokenURL    string = "https://www.googleapis.com/oauth2/v4/token"
	googleUserInfoURL string = "https://www.googleapis.com/oauth2/v1/userinfo"
)

type GoogleImpl struct {
	URLPrefix                *url.URL
	OAuthConfig              *config.OAuthConfiguration
	ProviderConfig           config.OAuthProviderConfiguration
	TimeProvider             coreTime.Provider
	UserInfoDecoder          UserInfoDecoder
	LoginIDNormalizerFactory loginid.LoginIDNormalizerFactory
}

func (f *GoogleImpl) GetAuthURL(state State, encodedState string) (string, error) {
	d, err := FetchOIDCDiscoveryDocument(http.DefaultClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return "", err
	}
	return d.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		URLPrefix:      f.URLPrefix,
		Nonce:          state.HashedNonce,
		EncodedState:   encodedState,
		ExtraParams: map[string]string{
			"prompt": "select_account",
		},
	}), nil
}

func (f *GoogleImpl) Type() config.OAuthProviderType {
	return config.OAuthProviderTypeGoogle
}

func (f *GoogleImpl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, state)
}

func (f *GoogleImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	d, err := FetchOIDCDiscoveryDocument(http.DefaultClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to get OIDC discovery document")
		return
	}
	// TODO(sso): Cache JWKs
	keySet, err := d.FetchJWKs(http.DefaultClient)
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to get OIDC JWKs")
		return
	}

	var tokenResp AccessTokenResp
	claims, err := d.ExchangeCode(
		http.DefaultClient,
		r.Code,
		keySet,
		f.URLPrefix,
		f.ProviderConfig.ClientID,
		f.ProviderConfig.ClientSecret,
		redirectURI(f.URLPrefix, f.ProviderConfig),
		state.HashedNonce,
		f.TimeProvider.NowUTC,
		&tokenResp,
	)
	if err != nil {
		return
	}

	// Verify the issuer
	// https://developers.google.com/identity/protocols/OpenIDConnect#validatinganidtoken
	iss, ok := claims["iss"].(string)
	if !ok {
		err = NewSSOFailed(SSOUnauthorized, "invalid iss")
		return
	}
	if iss != "https://accounts.google.com" && iss != "accounts.google.com" {
		err = NewSSOFailed(SSOUnauthorized, "invalid iss")
		return
	}

	// Ensure sub exists
	sub, ok := claims["sub"].(string)
	if !ok {
		err = NewSSOFailed(SSOUnauthorized, "no sub")
		return
	}

	email, _ := claims["email"].(string)
	if email != "" {
		normalizer := f.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyType("email"))
		email, err = normalizer.Normalize(email)
		if err != nil {
			return
		}
	}

	authInfo.ProviderConfig = f.ProviderConfig
	authInfo.ProviderRawProfile = claims
	authInfo.ProviderAccessTokenResp = tokenResp
	authInfo.ProviderUserInfo = ProviderUserInfo{
		ID:    sub,
		Email: email,
	}

	return
}

func (f *GoogleImpl) ExternalAccessTokenGetAuthInfo(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		urlPrefix:       f.URLPrefix,
		oauthConfig:     f.OAuthConfig,
		providerConfig:  f.ProviderConfig,
		accessTokenURL:  googleTokenURL,
		userProfileURL:  googleUserInfoURL,
		userInfoDecoder: f.UserInfoDecoder,
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ OAuthProvider                   = &GoogleImpl{}
	_ OpenIDConnectProvider           = &GoogleImpl{}
	_ ExternalAccessTokenFlowProvider = &GoogleImpl{}
)
