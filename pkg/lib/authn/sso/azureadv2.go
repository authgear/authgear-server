package sso

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Azureadv2Impl struct {
	Clock                    clock.Clock
	RedirectURL              RedirectURLProvider
	ProviderConfig           config.OAuthSSOProviderConfig
	Credentials              config.OAuthClientCredentialsItem
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (f *Azureadv2Impl) getOpenIDConfiguration() (*OIDCDiscoveryDocument, error) {
	// OPTIMIZE(sso): Cache OpenID configuration

	tenant := f.ProviderConfig.Tenant
	var endpoint string
	// Azure special tenant
	//
	// If the azure tenant is `organizations` or `common`,
	// the developer should make use of `before_user_create` and `before_identity_create` hook
	// to disallow any undesire identity.
	// The `raw_profile` of the identity is the ID Token claims.
	// Refer to https://docs.microsoft.com/en-us/azure/active-directory/develop/id-tokens
	// to see what claims the token could contain.
	//
	// For `organizations`, the user can be any user of any organizational AD.
	// Therefore the developer should have a whitelist of AD tenant IDs.
	// In the incoming hook, check if `tid` matches one of the entry of the whitelist.
	//
	// For `common`, in addition to the users from `organizations`, any Microsoft personal account
	// could be the user.
	// In case of personal account, the `tid` is "9188040d-6c67-4c5b-b112-36a304b66dad".
	// Therefore the developer should first check if `tid` indicates personal account.
	// If yes, apply their logic to disallow the user creation.
	// One very common example is to look at the claim `email`.
	// Use a email address parser to parse the email address.
	// Obtain the domain and check if the domain is whitelisted.
	// For example, if the developer only wants user from hotmail.com to create user,
	// ensure `tid` is "9188040d-6c67-4c5b-b112-36a304b66dad" and ensure `email`
	// is of domain `@hotmail.com`.

	// As of 2019-09-23, two special values are observed.
	// To discover these values, create a new client
	// and try different options.
	switch tenant {
	// Special value for any organizational AD
	case "organizations":
		endpoint = "https://login.microsoftonline.com/organizations/v2.0/.well-known/openid-configuration"
	// Special value for any organizational AD and personal accounts (Xbox etc)
	case "common":
		endpoint = "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration"
	default:
		endpoint = fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0/.well-known/openid-configuration", tenant)
	}

	return FetchOIDCDiscoveryDocument(http.DefaultClient, endpoint)
}

func (*Azureadv2Impl) Type() config.OAuthSSOProviderType {
	return config.OAuthSSOProviderTypeAzureADv2
}

func (f *Azureadv2Impl) Config() config.OAuthSSOProviderConfig {
	return f.ProviderConfig
}

func (f *Azureadv2Impl) GetAuthURL(param GetAuthURLParam) (string, error) {
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

func (f *Azureadv2Impl) GetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, param)
}

func (f *Azureadv2Impl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, param GetAuthInfoParam) (authInfo AuthInfo, err error) {
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

	oid, ok := claims["oid"].(string)
	if !ok {
		err = OAuthProtocolError.New("oid not found in ID Token")
		return
	}
	// For "Microsoft Account", email usually exists.
	// For "AD guest user", email usually exists because to invite an user, the inviter must provide email.
	// For "AD user", email never exists even one is provided in "Authentication Methods".
	email, _ := claims["email"].(string)
	if email != "" {
		normalizer := f.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypeEmail)
		email, err = normalizer.Normalize(email)
		if err != nil {
			return
		}
	}

	authInfo.ProviderConfig = f.ProviderConfig
	authInfo.ProviderRawProfile = claims
	authInfo.ProviderAccessTokenResp = tokenResp
	authInfo.ProviderUserInfo = ProviderUserInfo{
		ID:    oid,
		Email: email,
	}

	return
}

func (f *Azureadv2Impl) GetPrompt(prompt []string) []string {
	// Azureadv2 only support single value for prompt
	// the first supporting value in the list will be used
	// the usage of `none` is for checking existing authentication and/or consent
	// which doesn't fit auth ui case
	// ref: https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-auth-code-flow
	for _, p := range prompt {
		if p == "login" {
			return []string{"login"}
		} else if p == "consent" {
			return []string{"consent"}
		} else if p == "select_account" {
			return []string{"select_account"}
		}
	}
	return []string{}
}

var (
	_ OAuthProvider         = &Azureadv2Impl{}
	_ OpenIDConnectProvider = &Azureadv2Impl{}
)
