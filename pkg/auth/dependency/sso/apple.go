package sso

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/crypto"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

var appleOIDCConfig = OIDCDiscoveryDocument{
	JWKSUri:               "https://appleid.apple.com/auth/keys",
	TokenEndpoint:         "https://appleid.apple.com/auth/token",
	AuthorizationEndpoint: "https://appleid.apple.com/auth/authorize",
}

type AppleImpl struct {
	URLPrefix      *url.URL
	OAuthConfig    *config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
	TimeProvider   coreTime.Provider
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

func (f *AppleImpl) Type() config.OAuthProviderType {
	return config.OAuthProviderTypeApple
}

func (f *AppleImpl) GetAuthURL(state State, encodedState string) (string, error) {
	return appleOIDCConfig.MakeOAuthURL(OIDCAuthParams{
		ProviderConfig: f.ProviderConfig,
		URLPrefix:      f.URLPrefix,
		Nonce:          state.Nonce,
		EncodedState:   encodedState,
	}), nil
}

func (f *AppleImpl) GetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r, state)
}

func (f *AppleImpl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse, state State) (authInfo AuthInfo, err error) {
	if subtle.ConstantTimeCompare([]byte(state.Nonce), []byte(crypto.SHA256String(r.Nonce))) != 1 {
		err = NewSSOFailed(InvalidParams, "invalid sso state")
		return
	}

	keySet, err := appleOIDCConfig.FetchJWKs(http.DefaultClient)
	if err != nil {
		err = NewSSOFailed(NetworkFailed, "failed to get OIDC JWKs")
		return
	}

	clientSecret, err := f.createClientSecret()
	if err != nil {
		err = errors.Newf("failed to create client secret: %w", err)
		return
	}

	var tokenResp AccessTokenResp
	claims, err := appleOIDCConfig.ExchangeCode(
		http.DefaultClient,
		r.Code,
		keySet,
		f.URLPrefix,
		f.ProviderConfig.ClientID,
		clientSecret,
		redirectURI(f.URLPrefix, f.ProviderConfig),
		r.Nonce,
		f.TimeProvider.NowUTC,
		&tokenResp,
	)
	if err != nil {
		return
	}

	// Verify the issuer
	// https://developer.apple.com/documentation/signinwithapplerestapi/verifying_a_user
	// The exact spec is
	// Verify that the iss field contains https://appleid.apple.com
	// Therefore, we use strings.Contains here.
	iss, ok := claims["iss"].(string)
	if !ok {
		err = NewSSOFailed(SSOUnauthorized, "invalid iss")
		return
	}
	if !strings.Contains(iss, "https://appleid.apple.com") {
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
	_ OAuthProvider         = &AppleImpl{}
	_ OpenIDConnectProvider = &AppleImpl{}
)
