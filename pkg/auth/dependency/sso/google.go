package sso

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/clock"
)

const (
	googleOIDCDiscoveryDocumentURL string = "https://accounts.google.com/.well-known/openid-configuration"
)

type GoogleImpl struct {
	Clock                    clock.Clock
	RedirectURL              RedirectURLProvider
	ProviderConfig           config.OAuthSSOProviderConfig
	Credentials              config.OAuthClientCredentialsItem
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (f *GoogleImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	d, err := FetchOIDCDiscoveryDocument(http.DefaultClient, googleOIDCDiscoveryDocumentURL)
	if err != nil {
		return "", err
	}
	return d.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		RedirectURI:    f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		Nonce:          param.Nonce,
		State:          param.State,
		ExtraParams: map[string]string{
			"prompt": "select_account",
		},
	}), nil
}

func (f *GoogleImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeGoogle
}

func (f *GoogleImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, param)
}

func (f *GoogleImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
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
	jwtToken, err := d.ExchangeCode(
		http.DefaultClient,
		f.Clock,
		r.Code,
		keySet,
		f.ProviderConfig.ClientID,
		f.Credentials.ClientSecret,
		f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		param.Nonce,
		&tokenResp,
	)
	if err != nil {
		return
	}

	claims, err := jwtToken.AsMap(context.TODO())
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
