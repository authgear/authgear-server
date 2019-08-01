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
)

const (
	azureadv2ConfigurationURLFormat string = "https://login.microsoftonline.com/%s/.well-known/openid-configuration"
)

type Azureadv2Impl struct {
	OAuthConfig    config.OAuthConfiguration
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
	endpoint := fmt.Sprintf(azureadv2ConfigurationURLFormat, f.ProviderConfig.Tenant)

	// nolint: gosec
	resp, err := http.Get(endpoint)
	if err != nil {
		return
	}
	defer resp.Body.Close()
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
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get JWK keys: %d", resp.StatusCode)
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
		providerConfig: f.ProviderConfig,
		state:          NewState(params),
		baseURL:        c.AuthorizationEndpoint,
		responseMode:   "form_post",
		nonce:          params.State.Nonce,
	}
	return authURL(p)
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
		err = fmt.Errorf("invalid nonce")
		return
	}

	c, err := f.getOpenIDConfiguration()
	if err != nil {
		return
	}
	keySet, err := f.getKeys(c.JWKSUri)
	if err != nil {
		return
	}

	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("client_id", f.ProviderConfig.ClientID)
	body.Set("code", r.Code)
	body.Set("redirect_uri", redirectURI(f.OAuthConfig, f.ProviderConfig))
	body.Set("client_secret", f.ProviderConfig.ClientSecret)

	resp, err := http.PostForm(c.TokenEndpoint, body)
	if err != nil {
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
		var errorResp ErrorResp
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return
		}
		err = respToError(errorResp)
		return
	}

	idToken := tokenResp.IDToken()
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no key id")
		}
		if key := keySet.LookupKeyID(keyID); len(key) == 1 {
			return key[0].Materialize()
		}
		return nil, fmt.Errorf("unable to find key")
	}

	mapClaims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(idToken, mapClaims, keyFunc)
	if err != nil {
		return
	}

	if !mapClaims.VerifyAudience(f.ProviderConfig.ClientID, true) {
		err = fmt.Errorf("invalid audience")
		return
	}
	hashedNonce, ok := mapClaims["nonce"].(string)
	if !ok {
		err = fmt.Errorf("no nonce")
		return
	}
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = fmt.Errorf("invalid nonce")
		return
	}

	oid, ok := mapClaims["oid"].(string)
	if !ok {
		err = fmt.Errorf("cannot find oid")
		return
	}
	email, ok := mapClaims["email"].(string)
	if !ok {
		err = fmt.Errorf("cannot find email")
		return
	}

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
