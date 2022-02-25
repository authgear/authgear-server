package sso

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Azureadb2cImpl struct {
	Clock                        clock.Clock
	RedirectURL                  RedirectURLProvider
	ProviderConfig               config.OAuthSSOProviderConfig
	Credentials                  config.OAuthClientCredentialsItem
	StandardAttributesNormalizer StandardAttributesNormalizer
}

func (f *Azureadb2cImpl) getOpenIDConfiguration() (*OIDCDiscoveryDocument, error) {
	tenant := f.ProviderConfig.Tenant
	policy := f.ProviderConfig.Policy

	endpoint := fmt.Sprintf(
		"https://%s.b2clogin.com/%s.onmicrosoft.com/%s/v2.0/.well-known/openid-configuration",
		tenant,
		tenant,
		policy,
	)

	return FetchOIDCDiscoveryDocument(http.DefaultClient, endpoint)
}

func (f *Azureadb2cImpl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeAzureADB2C
}

func (f *Azureadb2cImpl) Config() config.OAuthSSOProviderConfig {
	return f.ProviderConfig
}

func (f *Azureadb2cImpl) GetAuthURL(param GetAuthURLParam) (string, error) {
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

func (f *Azureadb2cImpl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, param)
}

func (f *Azureadb2cImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
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
	if !ok || sub == "" {
		err = OAuthProtocolError.New("sub not found in ID Token")
		return
	}

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub

	stdAttrs, err := f.Extract(claims)
	if err != nil {
		return
	}
	authInfo.StandardAttributes = stdAttrs

	err = f.StandardAttributesNormalizer.Normalize(authInfo.StandardAttributes)
	if err != nil {
		return
	}

	return
}

func (f *Azureadb2cImpl) Extract(claims map[string]interface{}) (stdattrs.T, error) {
	// Here is the list of possible builtin claims.
	// city: free text
	// country: free text
	// jobTitle: free text
	// legalAgeGroupClassification: a enum with undocumented variants
	// postalCode: free text
	// state: free text
	// streetAddress: free text
	// newUser: true means the user signed up newly
	// oid: sub is identical to it by default.
	// emails: if non-empty, the first value corresponds to standard claim
	// name: correspond to standard claim
	// given_name: correspond to standard claim
	// family_name: correspond to standard claim

	extractString := func(input map[string]interface{}, output stdattrs.T, key string) {
		if value, ok := input[key].(string); ok && value != "" {
			output[key] = value
		}
	}

	out := stdattrs.T{}

	extractString(claims, out, stdattrs.Name)
	extractString(claims, out, stdattrs.GivenName)
	extractString(claims, out, stdattrs.FamilyName)

	var email string
	if ifaceSlice, ok := claims["emails"].([]interface{}); ok {
		for _, iface := range ifaceSlice {
			if str, ok := iface.(string); ok && str != "" {
				email = str
			}
		}
	}
	out[stdattrs.Email] = email

	return stdattrs.Extract(out, stdattrs.ExtractOptions{
		EmailRequired: *f.ProviderConfig.Claims.Email.Required,
	})
}

func (f *Azureadb2cImpl) GetPrompt(prompt []string) []string {
	// The only supported value is login.
	// See https://docs.microsoft.com/en-us/azure/active-directory-b2c/openid-connect
	for _, p := range prompt {
		if p == "login" {
			return []string{"login"}
		}
	}
	return []string{}
}

var (
	_ OAuthProvider         = &Azureadb2cImpl{}
	_ OpenIDConnectProvider = &Azureadb2cImpl{}
)
