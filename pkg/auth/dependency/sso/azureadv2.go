package sso

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

type Azureadv2Impl struct {
	URLPrefix                *url.URL
	OAuthConfig              *config.OAuthConfiguration
	ProviderConfig           config.OAuthProviderConfiguration
	TimeProvider             coreTime.Provider
	LoginIDNormalizerFactory loginid.LoginIDNormalizerFactory
}

func (f *Azureadv2Impl) getOpenIDConfiguration() (*OIDCDiscoveryDocument, error) {
	// TODO(sso): Cache OpenID configuration

	tenant := f.ProviderConfig.Tenant
	var endpoint string
	// GUIDE(sso): Azure special tenant
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

func (f *Azureadv2Impl) Type() config.OAuthProviderType {
	return config.OAuthProviderTypeAzureADv2
}

func (f *Azureadv2Impl) GetAuthURL(state State, encodedState string) (string, error) {
	c, err := f.getOpenIDConfiguration()
	if err != nil {
		return "", err
	}
	return c.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		URLPrefix:      f.URLPrefix,
		Nonce:          state.Nonce,
		EncodedState:   encodedState,
	}), nil
}

func (f *Azureadv2Impl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, state)
}

func (f *Azureadv2Impl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	if subtle.ConstantTimeCompare([]byte(state.Nonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = NewSSOFailed(InvalidParams, "invalid sso state")
		return
	}

	c, err := f.getOpenIDConfiguration()
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to get OIDC discovery document")
		return
	}
	// TODO(sso): Cache JWKs
	keySet, err := c.FetchJWKs(http.DefaultClient)
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to get OIDC JWKs")
		return
	}

	var tokenResp AccessTokenResp
	claims, err := c.ExchangeCode(
		http.DefaultClient,
		r.Code,
		keySet,
		f.URLPrefix,
		f.ProviderConfig.ClientID,
		f.ProviderConfig.ClientSecret,
		redirectURI(f.URLPrefix, f.ProviderConfig),
		r.Nonce,
		f.TimeProvider.NowUTC,
		&tokenResp,
	)
	if err != nil {
		return
	}

	oid, ok := claims["oid"].(string)
	if !ok {
		err = NewSSOFailed(SSOUnauthorized, "no oid")
		return
	}
	// For "Microsoft Account", email usually exists.
	// For "AD guest user", email usually exists because to invite an user, the inviter must provide email.
	// For "AD user", email never exists even one is provided in "Authentication Methods".
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
		ID:    oid,
		Email: email,
	}

	return
}

var (
	_ OAuthProvider         = &Azureadv2Impl{}
	_ OpenIDConnectProvider = &Azureadv2Impl{}
)
