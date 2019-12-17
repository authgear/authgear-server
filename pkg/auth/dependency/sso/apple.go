package sso

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

type AppleImpl struct {
	URLPrefix      *url.URL
	OAuthConfig    *config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
	TimeProvider   coreTime.Provider
}

func (f *AppleImpl) getKeys() (*jwk.Set, error) {
	// TODO(sso): Cache JWKs
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
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

func (f *AppleImpl) createClientSecret() (clientSecret string, err error) {
	// https://developer.apple.com/documentation/signinwithapplerestapi/generate_and_validate_tokens
	key, err := crypto.ParseAppleP8PrivateKey([]byte(f.ProviderConfig.ClientSecret))
	if err != nil {
		return
	}

	now := f.TimeProvider.NowUTC()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		Issuer:    f.ProviderConfig.TeamID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(5 * time.Minute).Unix(),
		Audience:  "https://appleid.apple.com",
		Subject:   f.ProviderConfig.ClientID,
	})
	token.Header["kid"] = f.ProviderConfig.KeyID

	clientSecret, err = token.SignedString(key)
	if err != nil {
		return
	}

	return
}

func (f *AppleImpl) GetAuthURL(state State, encodedState string) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		urlPrefix:      f.URLPrefix,
		providerConfig: f.ProviderConfig,
		encodedState:   encodedState,
		baseURL:        "https://appleid.apple.com/auth/authorize",
		responseMode:   "form_post",
		nonce:          state.Nonce,
	}
	return authURL(p)
}

func (f *AppleImpl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, state)
}

func (f *AppleImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	if subtle.ConstantTimeCompare([]byte(state.Nonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = NewSSOFailed(InvalidParams, "invalid sso state")
		return
	}

	keySet, err := f.getKeys()
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to get OIDC JWKs")
		return
	}

	clientSecret, err := f.createClientSecret()
	if err != nil {
		err = errors.Newf("failed to create client secret: %w", err)
		return
	}

	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("client_id", f.ProviderConfig.ClientID)
	body.Set("code", r.Code)
	body.Set("redirect_uri", redirectURI(f.URLPrefix, f.ProviderConfig))
	body.Set("client_secret", clientSecret)

	resp, err := http.PostForm("https://appleid.apple.com/auth/token", body)
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

	// https://developer.apple.com/documentation/signinwithapplerestapi/verifying_a_user
	// The following code verify the id token according to the docs.

	idToken := tokenResp.IDToken()
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, NewSSOFailed(SSOUnauthorized, "no kid")
		}
		if key := keySet.LookupKeyID(keyID); len(key) == 1 {
			return key[0].Materialize()
		}
		return nil, NewSSOFailed(SSOUnauthorized, "failed to find signing key")
	}

	mapClaims := jwt.MapClaims{}
	// Verify the signature
	_, err = jwt.ParseWithClaims(idToken, mapClaims, keyFunc)
	if err != nil {
		err = NewSSOFailed(SSOUnauthorized, "invalid JWT signature")
		return
	}

	// Verify the nonce
	hashedNonce, ok := mapClaims["nonce"].(string)
	if !ok {
		err = NewSSOFailed(InvalidParams, "no nonce")
		return
	}
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = NewSSOFailed(SSOUnauthorized, "invalid nonce")
		return
	}

	// Verify the issuer
	if !mapClaims.VerifyIssuer("https://appleid.apple.com", true) {
		err = NewSSOFailed(SSOUnauthorized, "invalid iss")
		return
	}

	// Verify the audience
	if !mapClaims.VerifyAudience(f.ProviderConfig.ClientID, true) {
		err = NewSSOFailed(SSOUnauthorized, "invalid aud")
		return
	}

	// Verify exp
	now := f.TimeProvider.NowUTC().Unix()
	if !mapClaims.VerifyExpiresAt(now, true) {
		err = NewSSOFailed(SSOUnauthorized, "invalid exp")
		return
	}

	// Ensure sub exists
	sub, ok := mapClaims["sub"].(string)
	if !ok {
		err = NewSSOFailed(SSOUnauthorized, "no sub")
		return
	}

	email, _ := mapClaims["email"].(string)

	authInfo.ProviderConfig = f.ProviderConfig
	authInfo.ProviderRawProfile = mapClaims
	authInfo.ProviderAccessTokenResp = tokenResp
	authInfo.ProviderUserInfo = ProviderUserInfo{
		ID:    sub,
		Email: email,
	}

	return
}

var (
	_ OAuthProvider         = &AppleImpl{}
	_ OpenIDConnectProvider = &AppleImpl{}
)
