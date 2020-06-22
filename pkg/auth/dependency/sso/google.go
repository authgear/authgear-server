package sso

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	googleOIDCDiscoveryDocumentURL string = "https://accounts.google.com/.well-known/openid-configuration"
)

type GoogleImpl struct {
	URLPrefix                *url.URL
	RedirectURLFunc          RedirectURLFunc
	OAuthConfig              *config.OAuthConfiguration
	ProviderConfig           config.OAuthProviderConfiguration
	Clock                    clock.Clock
	UserInfoDecoder          UserInfoDecoder
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (f *GoogleImpl) GetAuthURL(state State, encodedState string) (string, error) {
	d, err := FetchOIDCDiscoveryDocument(http.DefaultClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return "", err
	}
	return d.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		RedirectURI:    f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
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
		f.RedirectURLFunc(f.URLPrefix, f.ProviderConfig),
		state.HashedNonce,
		f.Clock.NowUTC,
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

var (
	_ OAuthProvider         = &GoogleImpl{}
	_ OpenIDConnectProvider = &GoogleImpl{}
)
