package sso

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type Azureadv2Impl struct {
	URLPrefix      *url.URL
	OAuthConfig    *config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

type azureadv2OpenIDConfiguration struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSUri               string `json:"jwks_uri"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

func (f *Azureadv2Impl) getOpenIDConfiguration() (c azureadv2OpenIDConfiguration, err error) {
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

	// nolint: gosec
	resp, err := http.Get(endpoint)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.Newf("unexpected status code: %d", resp.StatusCode)
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&c)
	if err != nil {
		return
	}
	return
}

func (f *Azureadv2Impl) getKeys(endpoint string) (*jwk.Set, error) {
	// TODO(sso): Cache JWKs
	// nolint: gosec
	resp, err := http.Get(endpoint)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("unexpected status code: %d", resp.StatusCode)
	}
	return jwk.Parse(resp.Body)
}

func (f *Azureadv2Impl) GetAuthURL(params GetURLParams) (string, error) {
	c, err := f.getOpenIDConfiguration()
	if err != nil {
		return "", err
	}
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		state:          NewState(params),
		baseURL:        c.AuthorizationEndpoint,
		responseMode:   "form_post",
		nonce:          params.State.Nonce,
	}
	return authURL(p)
}

func (f *Azureadv2Impl) EncodeState(state State) (encodedState string, err error) {
	return EncodeState(f.OAuthConfig.StateJWTSecret, state)
}

func (f *Azureadv2Impl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *Azureadv2Impl) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r)
}

func (f *Azureadv2Impl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	state, err := DecodeState(f.OAuthConfig.StateJWTSecret, r.State)
	if err != nil {
		return
	}

	if subtle.ConstantTimeCompare([]byte(state.Nonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = NewSSOFailed(InvalidParams, "invalid sso state")
		return
	}

	c, err := f.getOpenIDConfiguration()
	if err != nil {
		err = errors.Newf("failed to get OIDC discovery document: %w", err)
		return
	}
	keySet, err := f.getKeys(c.JWKSUri)
	if err != nil {
		err = errors.Newf("failed to get OIDC JWKS: %w", err)
		return
	}

	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("client_id", f.ProviderConfig.ClientID)
	body.Set("code", r.Code)
	body.Set("redirect_uri", redirectURI(f.URLPrefix, f.ProviderConfig))
	body.Set("client_secret", f.ProviderConfig.ClientSecret)

	resp, err := http.PostForm(c.TokenEndpoint, body)
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to connect authorization server")
		return
	}
	defer resp.Body.Close()

	var tokenResp AccessTokenResp
	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		if err != nil {
			return
		}
	} else {
		var errorResp oauthErrorResp
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return
		}
		err = errorResp.AsError()
		return
	}

	idToken := tokenResp.IDToken()
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("no key id")
		}
		if key := keySet.LookupKeyID(keyID); len(key) == 1 {
			return key[0].Materialize()
		}
		return nil, errors.New("unable to find key")
	}

	mapClaims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(idToken, mapClaims, keyFunc)
	if err != nil {
		err = errors.WithSecondaryError(
			NewSSOFailed(SSOUnauthorized, "unexpected authorization response"),
			err,
		)
		return
	}

	if !mapClaims.VerifyAudience(f.ProviderConfig.ClientID, true) {
		err = errors.WithSecondaryError(
			NewSSOFailed(SSOUnauthorized, "unexpected authorization response"),
			errors.New("invalid audience"),
		)
		return
	}
	hashedNonce, ok := mapClaims["nonce"].(string)
	if !ok {
		err = errors.WithSecondaryError(
			NewSSOFailed(SSOUnauthorized, "unexpected authorization response"),
			errors.New("no nonce"),
		)
		return
	}
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = errors.WithSecondaryError(
			NewSSOFailed(SSOUnauthorized, "unexpected authorization response"),
			errors.New("invalid nonce"),
		)
		return
	}

	oid, ok := mapClaims["oid"].(string)
	if !ok {
		err = errors.WithSecondaryError(
			NewSSOFailed(SSOUnauthorized, "unexpected authorization response"),
			errors.New("cannot find oid"),
		)
		return
	}
	// For "Microsoft Account", email usually exists.
	// For "AD guest user", email usually exists because to invite an user, the inviter must provide email.
	// For "AD user", email never exists even one is provided in "Authentication Methods".
	email, _ := mapClaims["email"].(string)

	authInfo.ProviderConfig = f.ProviderConfig
	authInfo.ProviderRawProfile = mapClaims
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
