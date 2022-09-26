package sso

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ADFSImpl struct {
	Clock                        clock.Clock
	RedirectURL                  RedirectURLProvider
	ProviderConfig               config.OAuthSSOProviderConfig
	Credentials                  config.OAuthSSOProviderCredentialsItem
	StandardAttributesNormalizer StandardAttributesNormalizer
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

	extracted, err := stdattrs.Extract(claims, stdattrs.ExtractOptions{})
	if err != nil {
		return
	}

	// Transform upn into preferred_username
	if _, ok := extracted[stdattrs.PreferredUsername]; !ok {
		extracted[stdattrs.PreferredUsername] = upn
	}
	// Transform upn into email
	if _, ok := extracted[stdattrs.Email]; !ok {
		if emailErr := (validation.FormatEmail{}).CheckFormat(upn); emailErr == nil {
			// upn looks like an email address.
			extracted[stdattrs.Email] = upn
		}
	}

	extracted, err = stdattrs.Extract(extracted, stdattrs.ExtractOptions{
		EmailRequired: *f.ProviderConfig.Claims.Email.Required,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = extracted

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub

	err = f.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func (f *ADFSImpl) GetPrompt(prompt []string) []string {
	// adfs only support prompt=login
	// ref: https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/operations/ad-fs-prompt-login
	for _, p := range prompt {
		if p == "login" {
			return []string{"login"}
		}
	}
	return []string{}
}

var (
	_ OAuthProvider         = &ADFSImpl{}
	_ OpenIDConnectProvider = &ADFSImpl{}
)
