package sso

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ADFSImpl struct {
	Clock                    clock.Clock
	RedirectURL              RedirectURLProvider
	ProviderConfig           config.OAuthSSOProviderConfig
	Credentials              config.OAuthClientCredentialsItem
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (*ADFSImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeADFS
}

func (f *ADFSImpl) Config() config.OAuthSSOProviderConfig {
	return f.ProviderConfig
}

func (f *ADFSImpl) getOpenIDConfiguration() (*OIDCDiscoveryDocument, error) {
	endpoint := f.ProviderConfig.DiscoveryDocumentEndpoint
	return FetchOIDCDiscoveryDocument(http.DefaultClient, endpoint)
}

func (f *ADFSImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	c, err := f.getOpenIDConfiguration()
	if err != nil {
		return "", err
	}
	return c.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		RedirectURI:    f.RedirectURL.SSOCallbackURL(f.ProviderConfig).String(),
		Nonce:          param.Nonce,
		State:          param.State,
		Prompt:         f.GetPrompt(param.Prompt),
	}), nil
}

func (f *ADFSImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, param)
}

func (f *ADFSImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	c, err := f.getOpenIDConfiguration()
	if err != nil {
		return
	}

	// OPTIMIZE(sso): Cache JWKs
	keySet, err := c.FetchJWKs(http.DefaultClient)
	if err != nil {
		return
	}

	var tokenResp AccessTokenResp
	jwtToken, err := c.ExchangeCode(
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

	sub, ok := claims["sub"].(string)
	if !ok {
		err = OAuthProtocolError.New("sub not found in ID token")
		return
	}

	// The upn claim is documented here.
	// https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/operations/configuring-alternate-login-id
	upn, ok := claims["upn"].(string)
	if !ok {
		err = OAuthProtocolError.New("upn not found in ID token")
		return
	}

	preferredUsername := upn

	var email string
	if emailErr := (validation.FormatEmail{}).CheckFormat(upn); emailErr == nil {
		// upn looks like an email address.
		normalizer := f.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypeEmail)
		email, err = normalizer.Normalize(upn)
		if err != nil {
			return
		}
	}

	authInfo.ProviderConfig = f.ProviderConfig
	authInfo.ProviderRawProfile = claims
	authInfo.ProviderAccessTokenResp = tokenResp
	authInfo.ProviderUserInfo = ProviderUserInfo{
		ID:                sub,
		Email:             email,
		PreferredUsername: preferredUsername,
	}

	return
}

func (f *ADFSImpl) GetPrompt(prompt []string) []string {
	return []string{}
}

var (
	_ OAuthProvider         = &ADFSImpl{}
	_ OpenIDConnectProvider = &ADFSImpl{}
)
