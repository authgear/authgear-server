package sso

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/adfs"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ADFSImpl struct {
	Clock                        clock.Clock
	ProviderConfig               oauthrelyingparty.ProviderConfig
	Credentials                  config.OAuthSSOProviderCredentialsItem
	StandardAttributesNormalizer StandardAttributesNormalizer
	HTTPClient                   OAuthHTTPClient
}

func (f *ADFSImpl) Config() oauthrelyingparty.ProviderConfig {
	return f.ProviderConfig
}

func (f *ADFSImpl) getOpenIDConfiguration() (*OIDCDiscoveryDocument, error) {
	endpoint := adfs.ProviderConfig(f.ProviderConfig).DiscoveryDocumentEndpoint()
	return FetchOIDCDiscoveryDocument(f.HTTPClient, endpoint)
}

func (f *ADFSImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
	c, err := f.getOpenIDConfiguration()
	if err != nil {
		return "", err
	}
	return c.MakeOAuthURL(AuthorizationURLParams{
		ClientID:     f.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        f.ProviderConfig.Scope(),
		ResponseType: ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		Prompt:       f.GetPrompt(param.Prompt),
		Nonce:        param.Nonce,
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
	keySet, err := c.FetchJWKs(f.HTTPClient)
	if err != nil {
		return
	}

	var tokenResp AccessTokenResp
	jwtToken, err := c.ExchangeCode(
		f.HTTPClient,
		f.Clock,
		r.Code,
		keySet,
		f.ProviderConfig.ClientID(),
		f.Credentials.ClientSecret,
		param.RedirectURI,
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

	emailRequired := f.ProviderConfig.EmailClaimConfig().Required()
	extracted, err = stdattrs.Extract(extracted, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
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
